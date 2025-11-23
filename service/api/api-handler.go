package api

import (
    "net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
    // Public routes
    rt.router.GET("/", rt.getHelloWorld)
    rt.router.GET("/liveness", rt.liveness)

    // Login (no auth)
    rt.router.POST("/session", rt.doLogin)

    // Authenticated routes
    rt.router.GET("/context", rt.wrap(rt.getContextReply))

    // --- USER ---
    rt.router.PUT("/users/username", rt.wrap(rt.setMyUserName))
    rt.router.GET("/users", rt.wrap(rt.listUsers))
    rt.router.PUT("/users/photo", rt.wrap(rt.setMyPhoto))

    // --- CONVERSATIONS ---
    rt.router.GET("/conversations", rt.wrap(rt.getMyConversations))
    rt.router.GET("/conversations/:chatId", rt.wrap(rt.getConversation))

    // --- MESSAGES ---
    rt.router.POST("/messages", rt.wrap(rt.sendMessage))
    rt.router.POST("/messages/:messageId/forward", rt.wrap(rt.forwardMessage))
    rt.router.POST("/messages/:messageId/reactions", rt.wrap(rt.commentMessage))
    rt.router.DELETE("/messages/:messageId/reactions/:emoji", rt.wrap(rt.uncommentMessage))
    rt.router.DELETE("/messages/:messageId", rt.wrap(rt.deleteMessage))

    // --- GROUPS ---
    rt.router.POST("/groups", rt.wrap(rt.createGroup))
    rt.router.POST("/groups/:chatId/members", rt.wrap(rt.addToGroup))
    rt.router.DELETE("/groups/:chatId/members/me", rt.wrap(rt.leaveGroup))
    rt.router.PUT("/groups/:chatId/name", rt.wrap(rt.setGroupName))
    rt.router.PUT("/groups/:chatId/photo", rt.wrap(rt.setGroupPhoto))

    return rt.router
}

