package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"nimadir/internal/db"
	"nimadir/internal/summarizer"
	"strings"
)

var telegramToken = "7525155845:AAFKu9pvUUPd4QcvlOa5cwY8MrMagBD8I48"
var geminiAPIKey = "AIzaSyCBQ3c-xPQNOy9joKbF9g0_OEzYgPDUVzw"

func StartBot() error {
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		return fmt.Errorf("failed to initialize bot: %v", err)
	}

	log.Println("Bot initialized successfully")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID

		if update.Message.IsCommand() {
			if update.Message.Command() == "start" {
				user, err := db.GetUserByChatID(chatID)
				if err != nil {
					log.Printf("Error fetching user by chatID %d: %v", chatID, err)
					sendMarkdownMessage(bot, chatID, "⚠️ *Xatolik yuz berdi, iltimos keyinroq urinib ko‘ring\\.*")
					continue
				}

				msg := tgbotapi.NewMessage(chatID, "")
				msg.ParseMode = "MarkdownV2"
				if user == nil {
					err = db.CreateUser(chatID, update.Message.From.UserName, "Free")
					if err != nil {
						log.Printf("Error creating user for chatID %d: %v", chatID, err)
						sendMarkdownMessage(bot, chatID, "⚠️ *Foydalanuvchi yaratishda xatolik yuz berdi\\.*")
						continue
					}
					msg.Text = "👋 *Salom\\!* Do‘stona AI yordamchingizga xush kelibsiz 🤖\\. Menga matn yuboring va xulosani oling\\!"
				} else {
					msg.Text = "👋 *Qaytganingizdan xursandman\\!* 😊\\. Mendan qanday foydalanishni o'zingiz juda yaxshi bilasiz\\!"
				}
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Error sending start message to chatID %d: %v", chatID, err)
					continue
				}

				menu := tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("📂 Profil"),
						tgbotapi.NewKeyboardButton("❓ Yordam"),
						tgbotapi.NewKeyboardButton("🗞 Tariflar"),
					),
				)
				menuMsg := tgbotapi.NewMessage(chatID, "📍 *Menyu:*")
				menuMsg.ReplyMarkup = menu
				menuMsg.ParseMode = "MarkdownV2"
				if _, err := bot.Send(menuMsg); err != nil {
					log.Printf("Error sending menu to chatID %d: %v", chatID, err)
					continue
				}
			}
			continue
		}

		switch update.Message.Text {
		case "📂 Profil":
			user, err := db.GetUserByChatID(chatID)
			if err != nil {
				log.Printf("Error fetching user for profile, chatID %d: %v", chatID, err)
				sendMarkdownMessage(bot, chatID, "⚠️ *Foydalanuvchi ma'lumotlarini olishda xatolik yuz berdi\\.*")
				continue
			}
			if user != nil {
				username := escapeMarkdownV2(user.Username)
				text := fmt.Sprintf("👤 *Username:* @%s\n📦 *Tarif:* %s", username, escapeMarkdownV2(user.Tariff))
				sendMarkdownMessage(bot, chatID, text)
			}
			continue
		case "❓ Yordam":
			sendMarkdownMessage(bot, chatID, "✍️ *Savolingizni yozib qoldiring:* @nimadir_321")
			continue
		case "🗞 Tariflar":
			btn := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonURL("🌐 *Tariflarni Ko‘rish*", "https://youtube.com"),
				),
			)
			msg := tgbotapi.NewMessage(chatID, "📦 *Tarif rejalar bilan tanishish uchun tugmani bosing:*")
			msg.ReplyMarkup = btn
			msg.ParseMode = "MarkdownV2"
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error sending tariffs message to chatID %d: %v", chatID, err)
				continue
			}
			continue
		}

		user, err := db.GetUserByChatID(chatID)
		if err != nil || user == nil {
			log.Printf("Error fetching user or user not found for chatID %d: %v", chatID, err)
			sendMarkdownMessage(bot, chatID, "⚠️ *Foydalanuvchi topilmadi yoki xatolik yuz berdi\\.*")
			continue
		}

		if summarizer.IsOverLimit(user.Tariff, update.Message.Text) {
			sendMarkdownMessage(bot, chatID, "⚠️ *Matn uzunligi tarif limitidan oshib ketdi\\!*\n📏 *Iltimos, qisqaroq matn yuboring\\.*")
			continue
		}

		limitedText := summarizer.LimitText(user.Tariff, update.Message.Text)
		waitMsg := tgbotapi.NewMessage(chatID, "⏳ *Javob tayyorlanmoqda...*")
		waitMsg.ParseMode = "MarkdownV2"
		sentMsg, err := bot.Send(waitMsg)
		if err != nil {
			log.Printf("Error sending wait message to chatID %d: %v", chatID, err)
			continue
		}

		result, err := summarizer.Summarize(geminiAPIKey, limitedText)
		if err != nil {
			log.Printf("Error summarizing text for chatID %d: %v", chatID, err)
			sendMarkdownMessage(bot, chatID, "⚠️ *Xatolik yuz berdi, keyinroq urinib ko‘ring\\.*")
			continue
		}

		result = escapeMarkdownV2(result)
		edit := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, "💬 *"+result+"*")
		edit.ParseMode = "MarkdownV2"
		if _, err := bot.Send(edit); err != nil {
			log.Printf("Error editing message for chatID %d: %v", sentMsg.MessageID, err)
			continue
		}
	}
	return nil
}

// sendMarkdownMessage sends a message with MarkdownV2 formatting
func sendMarkdownMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "MarkdownV2"
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending message to chatID %d: %v", chatID, err)
	}
}

// escapeMarkdownV2 escapes special characters for Telegram's MarkdownV2
func escapeMarkdownV2(text string) string {
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, char := range specialChars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}
	return text
}
