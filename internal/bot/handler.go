package bot

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"
	"time"

	"telegram_chabot/internal/db"
)

type Handler struct {
	DB *sql.DB
}

var awaitingTrustedAdd = make(map[int64]bool)
var awaitingTrustedRemove = make(map[int64]bool)

func (h *Handler) HandleUpdate(raw []byte) {
	log.Println("HANDLE UPDATE CALLED")

	var update Update
	if err := json.Unmarshal(raw, &update); err != nil {
		log.Println("Failed to parse update:", err)
		return
	}
	if update.Message == nil || update.Message.From == nil || update.Message.Chat == nil {
		return
	}

	msg := update.Message
	tgUser := msg.From

	u, err := db.GetOrCreateUser(
		h.DB,
		tgUser.ID,
		tgUser.Username,
		tgUser.FirstName,
		tgUser.LastName,
	)

	if awaitingTrustedAdd[u.ID] && msg.ForwardFrom != nil {
		contact := msg.ForwardFrom

		err := db.AddTrustedContact(
			h.DB,
			u.ID,
			contact.ID,
			contact.FirstName,
		)
		if err != nil {
			SendMessage(msg.Chat.ID, "⚠️ Failed to add trusted contact.")
			return
		}

		delete(awaitingTrustedAdd, u.ID)

		SendMessage(
			msg.Chat.ID,
			"✅ "+contact.FirstName+" has been added to your trusted contacts.",
		)
		return
	}

	if awaitingTrustedRemove[u.ID] && msg.ForwardFrom != nil {
		contact := msg.ForwardFrom

		err := db.RemoveTrustedContact(
			h.DB,
			u.ID,
			contact.ID,
		)
		if err != nil {
			SendMessage(msg.Chat.ID, "⚠️ Failed to remove trusted contact.")
			return
		}

		delete(awaitingTrustedRemove, u.ID)

		SendMessage(
			msg.Chat.ID,
			"❌ "+contact.FirstName+" has been removed from your trusted contacts.",
		)
		return
	}

	if awaitingTrustedAdd[u.ID] && msg.ForwardFrom != nil {
		contact := msg.ForwardFrom

		err := db.AddTrustedContact(
			h.DB,
			u.ID,
			contact.ID,
			contact.FirstName,
		)

		if awaitingTrustedRemove[u.ID] && msg.ForwardFrom != nil {
			contact := msg.ForwardFrom

			err := db.RemoveTrustedContact(
				h.DB,
				u.ID,
				contact.ID,
			)
			if err != nil {
				SendMessage(msg.Chat.ID, "⚠️ Failed to remove trusted contact.")
				return
			}

			delete(awaitingTrustedRemove, u.ID)

			SendMessage(
				msg.Chat.ID,
				"❌ "+contact.FirstName+" has been removed from your trusted contacts.",
			)
			return
		}

		if err != nil {
			SendMessage(msg.Chat.ID, "⚠️ Failed to add trusted contact.")
			return
		}

		delete(awaitingTrustedAdd, u.ID)

		SendMessage(
			msg.Chat.ID,
			"✅ "+contact.FirstName+" can now request your driving status.",
		)
		return
	}

	if err != nil {
		log.Println("GetOrCreateUser error:", err)
		return
	}

	if msg.Location != nil {
		var liveUntil *time.Time
		if msg.LivePeriod != nil && *msg.LivePeriod > 0 {
			t := time.Now().Add(time.Duration(*msg.LivePeriod) * time.Second)
			liveUntil = &t
		}

		_ = db.UpsertUserLocation(
			h.DB,
			u.ID,
			msg.Location.Latitude,
			msg.Location.Longitude,
			liveUntil,
		)

		SendMessage(msg.Chat.ID, "📍 Location saved.")
		return
	}

	text := strings.TrimSpace(msg.Text)
	normalized := NormalizeText(text)

	if strings.HasPrefix(text, "/") {
		h.handleCommand(u.ID, msg.Chat.ID, text)
		return
	}

	intent := DetectModelAIntent(normalized)

	log.Println("MODEL A INTENT:", intent, "TEXT:", normalized)

	switch intent {
	case IntentDrivingStatusRequest:
		h.handleDrivingStatusRequest(u.ID, msg.Chat.ID, tgUser.ID, normalized)

		return

	case IntentConfirmShare:
		h.handleConfirmShare(msg.Chat.ID, tgUser.ID)
		return

	case IntentCancelRequest:
		_ = db.CancelPendingRequestsForRequester(h.DB, tgUser.ID)
		SendMessage(msg.Chat.ID, "❌ Request cancelled.")
		return
	}

	// 5) Fallback to rule engine
	h.evaluateAndExecute(u.ID, msg.Chat.ID, tgUser.ID, text)
}

func (h *Handler) handleCommand(userID int64, chatID int64, cmd string) {
	switch cmd {

	case "/start":
		SendMessage(
			chatID,
			"👋 Welcome to Safety Assistant Bot\n\n"+
				"This bot helps share your location safely while driving.\n\n"+
				"🚗 How it works:\n"+
				"• Turn driving mode ON when driving\n"+
				"• Share your live location once\n"+
				"• Trusted contacts can ask if you’re driving\n"+
				"• Location is shared only after confirmation\n\n"+
				"Trusted contacts:\n"+
				"/add_trusted – allow someone to request your driving status\n"+
				"/remove_trusted – revoke access\n\n"+
				" Commands:\n"+
				"/driving_on\n"+
				"/driving_off\n"+
				"/status\n"+
				"/help",
		)
	case "/driving_on":
		_ = db.SetContext(h.DB, userID, "driving", true)
		SendMessage(chatID, "🚗 Driving mode ON")

	case "/driving_off":
		_ = db.SetContext(h.DB, userID, "driving", false)
		SendMessage(chatID, "🛑 Driving mode OFF")

	case "/status":
		h.sendStatus(userID, chatID)

	case "/help":
		SendMessage(
			chatID,
			"/driving_on – enable driving automation\n"+
				"/driving_off – disable driving automation\n"+
				"/status – show current automation status\n\n"+
				"Tip: Share a LIVE location once to enable automatic replies."+
				"To check if someone is driving, just ask: is [name] driving?",
		)
	case "/add_trusted":
		awaitingTrustedAdd[userID] = true
		SendMessage(
			chatID,
			" Please forward me a message from the person you want to trust.",
		)
	case "/remove_trusted":
		awaitingTrustedRemove[userID] = true
		SendMessage(
			chatID,
			" Please forward me a message from the person you want to remove.",
		)

	default:
		SendMessage(chatID, "Unknown command. Use /help")
	}
}

func (h *Handler) sendStatus(userID int64, chatID int64) {
	active, _ := db.IsContextActive(h.DB, userID, "driving")

	lat, lon, liveUntil, ok, _ := db.GetUserLocation(h.DB, userID)

	status := "🛑 Driving mode: OFF\n"
	if active {
		status = "🚗 Driving mode: ON\n"
	}

	if !ok {
		status += "📍 Location: NOT SET"
		SendMessage(chatID, status)
		return
	}

	if liveUntil != nil && liveUntil.After(time.Now()) {
		remaining := time.Until(*liveUntil).Round(time.Minute)
		status += "📍 Live location: ACTIVE (" + remaining.String() + " left)"
	} else {
		_ = lat
		_ = lon
		status += "📍 Location: SAVED (static)"
	}

	SendMessage(chatID, status)
}

func (h *Handler) evaluateAndExecute(
	userID int64,
	chatID int64,
	senderTelegramID int64,
	text string,
) {

	normalized := NormalizeText(text)

	rules, err := db.GetActiveRules(h.DB, userID)
	if err != nil || len(rules) == 0 {
		return
	}

	for _, r := range rules {

		// Context gate
		if r.RequiredContext.Valid {
			active, err := db.IsContextActive(h.DB, userID, r.RequiredContext.String)
			if err != nil || !active {
				continue
			}
		}

		ok, reason := MatchTriggers(h.DB, r.ID, normalized)
		if !ok {
			continue
		}

		if r.ApprovalMode != "auto" {
			_ = db.InsertActionLog(
				h.DB,
				userID,
				&r.ID,
				nil,
				text,
				"manual_required",
				"skipped",
				reason,
			)
			SendMessage(chatID, "⚠️ Action requires manual approval (not enabled).")
			return
		}

		trusted, err := db.IsTrustedContact(h.DB, userID, senderTelegramID)
		if err != nil {
			log.Println("Trusted contact check failed:", err)
			return
		}
		if !trusted {

			return
		}

		actions, err := db.GetRuleActions(h.DB, r.ID)
		if err != nil || len(actions) == 0 {
			return
		}

		executed := h.executeActions(userID, chatID, actions)
		status := "failed"
		if executed {
			status = "executed"
		}

		_ = db.InsertActionLog(
			h.DB,
			userID,
			&r.ID,
			nil,
			text,
			"rule_actions",
			status,
			reason,
		)

		return
	}
}

func (h *Handler) executeActions(userID int64, chatID int64, actions []db.RuleAction) bool {
	for _, a := range actions {
		switch a.Type {

		case "send_text":
			txt, _ := a.Payload["text"].(string)
			if txt == "" {
				txt = " Done."
			}
			SendMessage(chatID, txt)

		case "send_location", "send_live_location":
			lat, lon, liveUntil, ok, err := db.GetUserLocation(h.DB, userID)
			if err != nil || !ok {
				SendMessage(chatID, "❗ Please share your location once.")
				return false
			}

			now := time.Now()
			if liveUntil != nil && liveUntil.After(now) {
				secondsLeft := int(liveUntil.Sub(now).Seconds())
				if secondsLeft < 60 {
					secondsLeft = 60
				}
				SendLiveLocation(chatID, lat, lon, secondsLeft)
			} else {
				SendLocation(chatID, lat, lon)
			}

		default:
			log.Println("Unknown action type:", a.Type)
		}
	}
	return true
}

func (h *Handler) handleDrivingStatusRequest(
	requesterUserID int64,
	chatID int64,
	requesterTelegramID int64,
	normalizedText string,
) {

	mentionedName := ExtractName(normalizedText)

	if mentionedName != "" {
		targetUserID, found, err := db.FindUserByFirstName(h.DB, mentionedName)
		if err != nil {
			log.Println("FindUserByFirstName error:", err)
			return
		}

		if !found {
			SendMessage(chatID, "⚠️ I couldn’t find anyone named "+mentionedName+".")
			return
		}

		trusted, _ := db.IsTrustedContact(h.DB, targetUserID, requesterTelegramID)
		if !trusted {
			SendMessage(chatID, "⚠️ You are not in "+mentionedName+"’s trusted contacts.")
			return
		}

		active, _ := db.IsContextActive(h.DB, targetUserID, "driving")
		if !active {
			SendMessage(chatID, "🛑 No, "+mentionedName+" is not currently driving.")
			return
		}

		expires := time.Now().Add(5 * time.Minute)
		_, _ = db.CreatePendingRequest(
			h.DB,
			targetUserID,
			requesterTelegramID,
			chatID,
			expires,
		)

		_ = db.InsertActionLog(
			h.DB,
			targetUserID,
			nil,
			nil,
			"is driving?",
			"request_created",
			"pending",
			"explicit_name",
		)

		SendMessage(
			chatID,
			"🚗 Yes, "+mentionedName+" is currently driving.\nDo you want me to share their location?",
		)
		return
	}

	owners, err := db.GetOwnersForRequester(h.DB, requesterTelegramID)
	if err != nil {
		log.Println("GetOwnersForRequester error:", err)
		return
	}

	if len(owners) == 0 {
		SendMessage(chatID, "⚠️ I can’t help with that request.")
		return
	}

	var activeOwner int64
	for _, ownerID := range owners {
		active, _ := db.IsContextActive(h.DB, ownerID, "driving")
		if active {
			if activeOwner != 0 {
				SendMessage(chatID, "⚠️ Please specify who you are asking about.")
				return
			}
			activeOwner = ownerID
		}
	}

	if activeOwner == 0 {
		SendMessage(chatID, "🛑 No, they are not currently driving.")
		return
	}

	ownerName := "this person"
	if name, err := db.GetUserDisplayName(h.DB, activeOwner); err == nil {
		ownerName = name
	}

	expires := time.Now().Add(5 * time.Minute)
	_, _ = db.CreatePendingRequest(
		h.DB,
		activeOwner,
		requesterTelegramID,
		chatID,
		expires,
	)

	_ = db.InsertActionLog(
		h.DB,
		activeOwner,
		nil,
		nil,
		"is driving?",
		"request_created",
		"pending",
		"inferred_owner",
	)

	SendMessage(
		chatID,
		"🚗 Yes, "+ownerName+" is currently driving.\nDo you want me to share their location?",
	)
}

func (h *Handler) handleConfirmShare(
	chatID int64,
	requesterTelegramID int64,
) {
	pr, err := db.GetLatestPendingRequest(h.DB, requesterTelegramID)
	if err != nil {
		log.Println("GetLatestPendingRequest error:", err)
		return
	}

	if pr == nil {
		SendMessage(chatID, "⏱️ That request has expired. Please ask again.")
		return
	}

	ownerName := "this person"
	if name, err := db.GetUserDisplayName(h.DB, pr.OwnerUserID); err == nil {
		ownerName = name
	}

	trusted, _ := db.IsTrustedContact(h.DB, pr.OwnerUserID, requesterTelegramID)
	if !trusted {
		SendMessage(chatID, "⚠️ "+ownerName+" has not enabled location sharing with you.")
		_ = db.InsertActionLog(
			h.DB,
			pr.OwnerUserID,
			nil,
			nil,
			"location share",
			"request_denied",
			"blocked",
			"not trusted",
		)
		return
	}

	lat, lon, liveUntil, ok, _ := db.GetUserLocation(h.DB, pr.OwnerUserID)
	if !ok {
		SendMessage(chatID, "⚠️ "+ownerName+"'s location is not available.")
		return
	}

	SendMessage(chatID, "📍 Here is "+ownerName+"’s location:")

	now := time.Now()
	if liveUntil != nil && liveUntil.After(now) {
		seconds := int(liveUntil.Sub(now).Seconds())
		if seconds < 60 {
			seconds = 60
		}
		SendLiveLocation(chatID, lat, lon, seconds)
	} else {
		SendLocation(chatID, lat, lon)
	}

	_ = db.MarkPendingRequestCompleted(h.DB, pr.ID)

	_ = db.InsertActionLog(
		h.DB,
		pr.OwnerUserID,
		nil,
		nil,
		"location share",
		"request_confirmed",
		"executed",
		"",
	)
}
