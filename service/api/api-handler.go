package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	const base = "/v1"

	// Public routes
	rt.router.GET(base+"/", rt.getHelloWorld)
	rt.router.GET(base+"/liveness", rt.liveness)
	rt.router.ServeFiles(base+"/uploads/*filepath", http.Dir("./uploads"))

	// Login (no auth)
	rt.router.POST(base+"/session", rt.doLogin)

	// Authenticated routes
	rt.router.GET(base+"/context", rt.wrap(rt.getContextReply))

	// --- USER ---
	rt.router.PUT(base+"/users/username", rt.wrap(rt.setMyUserName))
	rt.router.GET(base+"/users", rt.wrap(rt.listUsers))
	rt.router.PUT(base+"/users/photo", rt.wrap(rt.setMyPhoto))

	// --- CONVERSATIONS ---
	rt.router.GET(base+"/conversations", rt.wrap(rt.listMyConversations))
	rt.router.GET(base+"/conversations/:chatId", rt.wrap(rt.listConversationMessages))
	
	rt.router.POST(base+"/conversations/:chatId/received", rt.wrap(rt.markConversationReceived))
	rt.router.POST(base+"/conversations/:chatId/read", rt.wrap(rt.markConversationRead))

	// --- MESSAGES ---
	rt.router.POST(base+"/messages", rt.wrap(rt.sendMessage))
	rt.router.POST(base+"/messages/:messageId/forward", rt.wrap(rt.forwardMessage))
	rt.router.POST(base+"/messages/:messageId/reactions", rt.wrap(rt.commentMessage))
	rt.router.DELETE(base+"/messages/:messageId/reactions/:emoji", rt.wrap(rt.uncommentMessage))
	rt.router.DELETE(base+"/messages/:messageId", rt.wrap(rt.deleteMessage))

	// --- DIRECT ---
	rt.router.POST(base+"/direct", rt.wrap(rt.createDirect))

	// --- GROUPS ---
	rt.router.POST(base+"/groups", rt.wrap(rt.createGroup))
	rt.router.POST(base+"/groups/:chatId/members", rt.wrap(rt.addToGroup))
	rt.router.DELETE(base+"/groups/:chatId/members/me", rt.wrap(rt.leaveGroup))
	rt.router.PUT(base+"/groups/:chatId/name", rt.wrap(rt.setGroupName))
	rt.router.PUT(base+"/groups/:chatId/photo", rt.wrap(rt.setGroupPhoto))

	// --- MEDIA ---
	rt.router.POST(base+"/media", rt.wrap(rt.uploadMedia))

	return rt.router
}
