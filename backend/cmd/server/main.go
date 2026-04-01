package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"wechat-enterprise-backend/internal/config"
	"wechat-enterprise-backend/internal/db"
	"wechat-enterprise-backend/internal/handler"
	"wechat-enterprise-backend/internal/middleware"
	"wechat-enterprise-backend/internal/realtime"
	"wechat-enterprise-backend/internal/service"
	"wechat-enterprise-backend/internal/wechat"
	"wechat-enterprise-backend/pkg/jwt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	gdb, err := db.Connect(cfg.MySQL)
	if err != nil {
		log.Fatalf("connect mysql failed: %v", err)
	}

	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.ExpireHours)
	authService := service.NewAuthService(gdb, jwtManager)
	if err := authService.EnsureSeedAdmin(context.Background(), cfg.Seed.AdminUsername, cfg.Seed.AdminPassword, cfg.Seed.AdminName); err != nil {
		log.Fatalf("seed admin failed: %v", err)
	}

	wechatClient := wechat.NewClient(cfg.Wechat.BaseURL)
	accountService := service.NewAccountService(gdb, wechatClient, cfg.Wechat.DefaultLoginPlatform, cfg.AI)
	realtimeHub := realtime.NewHub()

	authHandler := handler.NewAuthHandler(authService)
	accountHandler := handler.NewAccountHandler(accountService)
	realtimeHandler := handler.NewRealtimeHandler(accountService, realtimeHub)

	if cfg.AppEnv != "production" {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.Default()
	corsConfig := cors.Config{
		AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Authorization", "Content-Type", "Origin", "Accept", "X-Requested-With"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}
	if cfg.AppEnv == "production" {
		corsConfig.AllowOriginWithContextFunc = func(_ *gin.Context, origin string) bool {
			return isAllowedDevOrigin(origin)
		}
		corsConfig.AllowCredentials = true
	} else {
		// 开发环境放开来源限制，避免本地多端口/局域网调试被 CORS 阻断。
		corsConfig.AllowOriginWithContextFunc = func(_ *gin.Context, _ string) bool {
			return true
		}
	}
	router.Use(cors.New(corsConfig))
	_ = router.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/ws/:wxid", realtimeHandler.ServeWS)
	msg := router.Group("/msg")
	{
		msg.POST("/SyncMessage/:wxid", realtimeHandler.SyncMessage)
		msg.POST("/SyncMessage1/:wxid", realtimeHandler.SyncMessagePreview)
	}

	api := router.Group("/api")
	{
		api.POST("/auth/login", authHandler.Login)

		protected := api.Group("")
		protected.Use(middleware.AuthRequired(jwtManager))
		{
			protected.GET("/auth/me", authHandler.Me)
			protected.GET("/dashboard/overview", accountHandler.DashboardOverview)
			protected.GET("/ai/providers", accountHandler.ListAIProviders)

			protected.GET("/accounts", accountHandler.ListAccounts)
			protected.POST("/accounts/login-sessions", accountHandler.CreateLoginSession)
			protected.GET("/accounts/login-sessions/:sessionId", accountHandler.PollLoginSession)
			protected.POST("/accounts/:wxid/awaken-login-sessions", accountHandler.CreateAwakenLoginSession)
			protected.POST("/accounts/:wxid/bootstrap", accountHandler.Bootstrap)
			protected.POST("/accounts/:wxid/heartbeat/start", accountHandler.StartHeartbeat)
			protected.POST("/accounts/:wxid/heartbeat/stop", accountHandler.StopHeartbeat)
			protected.POST("/accounts/:wxid/logout", accountHandler.Logout)
			protected.GET("/accounts/:wxid/contacts", accountHandler.ListContacts)
			protected.POST("/accounts/:wxid/contacts/reload", accountHandler.ReloadContacts)
			protected.GET("/accounts/:wxid/finder/profile", accountHandler.GetFinderProfile)
			protected.GET("/accounts/:wxid/favorites", accountHandler.ListFavorites)
			protected.GET("/accounts/:wxid/favorites/:favId", accountHandler.GetFavoriteDetail)
			protected.POST("/accounts/:wxid/favorites/:favId/delete", accountHandler.DeleteFavorite)
			protected.GET("/accounts/:wxid/moments", accountHandler.ListMoments)
			protected.GET("/accounts/:wxid/moments/:momentId", accountHandler.GetMomentDetail)
			protected.POST("/accounts/:wxid/moments", accountHandler.PublishMoment)
			protected.POST("/accounts/:wxid/moments/:momentId/operation", accountHandler.OperateMoment)
			protected.POST("/accounts/:wxid/friends/search", accountHandler.SearchFriendCandidates)
			protected.POST("/accounts/:wxid/friends/request", accountHandler.SendFriendRequest)
			protected.POST("/accounts/:wxid/friends/request/batch", accountHandler.SendFriendRequestBatch)
			protected.POST("/accounts/:wxid/friends/:targetWxid/relation", accountHandler.CheckFriendRelation)
			protected.POST("/accounts/:wxid/friends/relation/batch", accountHandler.CheckFriendRelationBatch)
			protected.POST("/accounts/:wxid/friends/:targetWxid/delete", accountHandler.DeleteFriend)
			protected.POST("/accounts/:wxid/friends/:targetWxid/blacklist", accountHandler.SetFriendBlacklist)
			protected.POST("/accounts/:wxid/groups/:qid/refresh", accountHandler.RefreshGroup)
			protected.GET("/accounts/:wxid/groups/:qid/members", accountHandler.GetGroupMembers)
			protected.PUT("/accounts/:wxid/groups/:qid/name", accountHandler.UpdateGroupName)
			protected.PUT("/accounts/:wxid/groups/:qid/announcement", accountHandler.UpdateGroupAnnouncement)
			protected.PUT("/accounts/:wxid/groups/:qid/remark", accountHandler.UpdateGroupRemark)
			protected.POST("/accounts/:wxid/groups/:qid/address-book", accountHandler.SetGroupAddressBook)
			protected.POST("/accounts/:wxid/groups/:qid/members/add", accountHandler.AddGroupMembers)
			protected.POST("/accounts/:wxid/groups/:qid/members/invite", accountHandler.InviteGroupMembers)
			protected.POST("/accounts/:wxid/groups/:qid/members/remove", accountHandler.RemoveGroupMembers)
			protected.POST("/accounts/:wxid/groups/:qid/admin", accountHandler.OperateGroupAdmin)
			protected.POST("/accounts/:wxid/groups/:qid/add-friend", accountHandler.AddGroupFriend)
			protected.POST("/accounts/:wxid/groups/:qid/quit", accountHandler.QuitGroup)
			protected.GET("/accounts/:wxid/conversations", accountHandler.ListConversations)
			protected.GET("/accounts/:wxid/conversations/:conversationId", accountHandler.GetConversationDetail)
			protected.GET("/accounts/:wxid/conversations/:conversationId/ai-setting", accountHandler.GetConversationAISetting)
			protected.PUT("/accounts/:wxid/conversations/:conversationId/ai-setting", accountHandler.UpdateConversationAISetting)
			protected.POST("/accounts/:wxid/conversations/:conversationId/ai-draft", accountHandler.GenerateConversationAIDraft)
			protected.DELETE("/accounts/:wxid/conversations/:conversationId", accountHandler.DeleteConversation)
			protected.GET("/accounts/:wxid/conversations/:conversationId/messages", accountHandler.ListConversationMessages)
			protected.GET("/accounts/:wxid/conversations/:conversationId/messages/:messageId/image", accountHandler.DownloadConversationImage)
			protected.POST("/accounts/:wxid/conversations/:conversationId/messages/text", accountHandler.SendConversationText)
			protected.POST("/accounts/:wxid/conversations/:conversationId/messages/image", accountHandler.SendConversationImage)
			protected.POST("/accounts/:wxid/conversations/:conversationId/messages/emoji", accountHandler.SendConversationEmoji)
			protected.GET("/accounts/:wxid/messages/emojis", accountHandler.ListRecentConversationEmojis)
			protected.POST("/accounts/:wxid/conversations/:conversationId/messages/card", accountHandler.ShareConversationCard)
			protected.POST("/accounts/:wxid/conversations/:conversationId/messages/link", accountHandler.ShareConversationLink)
			protected.GET("/accounts/:wxid/messages", accountHandler.ListMessages)
			protected.POST("/accounts/:wxid/messages/sync", accountHandler.SyncMessages)
			protected.POST("/accounts/:wxid/messages/send-text", accountHandler.SendTextMessage)
		}
	}

	log.Printf("backend listening on :%s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}

func isAllowedDevOrigin(origin string) bool {
	if strings.TrimSpace(origin) == "" {
		return true
	}

	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}

	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return false
	}
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate()
}
