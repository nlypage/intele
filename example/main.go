package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/nlypage/intele/v2"
	"github.com/nlypage/intele/v2/storage"
	tele "gopkg.in/telebot.v3"
)

func main() {
	// Initialize Telegram bot
	bot, err := tele.NewBot(tele.Settings{
		Token:     "ENTER_YOUR_BOT_TOKEN",
		Poller:    &tele.LongPoller{Timeout: 10 * time.Second},
		ParseMode: tele.ModeMarkdown,
	})
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	// Create FlowBus with memory storage
	flowBus := intele.NewBus(bot, storage.NewMemoryStorage())

	// Register session restoration middleware
	bot.Use(flowBus.Restore())

	// Register flow message handlers
	bot.Handle(tele.OnText, flowBus.Handle)
	bot.Handle(tele.OnCallback, flowBus.Handle)
	bot.Handle(tele.OnContact, flowBus.Handle)
	bot.Handle(tele.OnLocation, flowBus.Handle)
	bot.Handle(tele.OnDocument, flowBus.Handle)
	bot.Handle(tele.OnPhoto, flowBus.Handle)

	// Setup dependency injection
	container := setupDependencies()

	// Create comprehensive user registration flow
	registrationFlow := createComprehensiveRegistrationFlow(flowBus, container)

	// Register commands
	bot.Handle("/start", func(c tele.Context) error {
		return c.Send(
			"🌟 **Welcome to Comprehensive Registration Demo**\n\n" +
				"This demonstrates all features of the Intele library:\n\n" +
				"✨ **Features included:**\n" +
				"• Multiple input types (text, numbers, contacts, files)\n" +
				"• Interactive callbacks and buttons\n" +
				"• Data validation and error handling\n" +
				"• Message collection and cleanup\n" +
				"• Session persistence and state management\n" +
				"• Conditional flow navigation\n" +
				"• Dependency injection\n" +
				"• Middleware support\n\n" +
				"Use /register to start the registration process\n" +
				"Use /cancel to cancel any active session\n" +
				"Use /status to check your current session",
		)
	})

	bot.Handle("/register", registrationFlow.Start)

	bot.Handle("/cancel", func(c tele.Context) error {
		userID := c.Sender().ID
		if err := flowBus.CancelSession(userID); err != nil {
			return c.Send("❌ No active session to cancel")
		}
		return c.Send("✅ Registration cancelled")
	})

	bot.Handle("/status", func(c tele.Context) error {
		userID := c.Sender().ID
		session, exists := flowBus.GetActiveSession(userID)
		if !exists {
			return c.Send("📭 No active registration session")
		}

		statusMessage := fmt.Sprintf(
			"📊 Registration Status\n\n"+
				"🏷️ Current Step: %s\n"+
				"⏰ Started: %s\n"+
				"🔄 Last Updated: %s\n"+
				"📝 Progress: %s",
			session.Data.CurrentStep,
			session.Data.CreatedAt.Format("15:04:05"),
			session.Data.UpdatedAt.Format("15:04:05"),
			getProgressText(session.Data.CurrentStep),
		)

		return c.Send(statusMessage, &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	log.Println("🚀 Comprehensive Registration Bot Started!")
	log.Println("Commands: /start, /register, /cancel, /status")
	bot.Start()
}

func setupDependencies() intele.Container {
	container := intele.NewContainer()
	container.Set("userService", &UserService{users: make(map[int64]*User)})
	container.Set("validator", &ValidationService{})
	container.Set("logger", &Logger{})
	return container
}

func createComprehensiveRegistrationFlow(flowBus *intele.FlowBus, container intele.Container) *intele.Flow {
	flow, err := flowBus.NewFlow("comprehensive_registration").
		Steps(
			// 1. Welcome Step
			intele.NewStep("welcome", func(c tele.Context, ctrl intele.Controller) error {
				collector := ctrl.MessageCollector()
				if c.Message() != nil {
					_ = collector.Collect(c.Message().ID, c.Message().Chat.ID)
				}

				markup := &tele.ReplyMarkup{}
				startBtn := markup.Data("🚀 Start Registration", "start_reg")
				infoBtn := markup.Data("ℹ️ More Info", "more_info")
				cancelBtn := markup.Data("❌ Cancel", "cancel_reg")

				markup.Inline(
					markup.Row(startBtn),
					markup.Row(infoBtn),
					markup.Row(cancelBtn),
				)

				return collector.Send(
					"👋 **Welcome to Comprehensive User Registration!**\n\n"+
						"This registration process will collect various types of information:\n\n"+
						"📝 **Personal Information**\n"+
						"📞 **Contact Details**\n"+
						"🏠 **Location Information**\n"+
						"💼 **Professional Background**\n"+
						"🎯 **Skills & Interests**\n"+
						"📊 **Experience Rating**\n"+
						"📎 **Document Upload**\n\n"+
						"Ready to begin?",
					markup,
				)
			}).
				AssignCallback("start_reg", "more_info", "cancel_reg").
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					if c.Callback() == nil {
						return nil
					}

					switch c.Callback().Unique {
					case "start_reg":
						return ctrl.Jump("basic_info")
					case "more_info":
						return ctrl.Jump("info_detail")
					case "cancel_reg":
						return ctrl.Cancel()
					}
					return nil
				}),

			// 2. Information Detail Step
			intele.NewStep("info_detail", func(c tele.Context, ctrl intele.Controller) error {
				collector := ctrl.MessageCollector()

				markup := &tele.ReplyMarkup{}
				backBtn := markup.Data("⬅️ Back", "back_to_welcome")
				startBtn := markup.Data("🚀 Start Now", "start_from_info")

				markup.Inline(
					markup.Row(startBtn),
					markup.Row(backBtn),
				)

				return collector.Send(
					"ℹ️ **Registration Process Overview**\n\n"+
						"**Step 1:** Basic Information (name, age, email)\n"+
						"**Step 2:** Contact preferences\n"+
						"**Step 3:** Location sharing (optional)\n"+
						"**Step 4:** Professional background\n"+
						"**Step 5:** Skills & interests selection\n"+
						"**Step 6:** Experience rating\n"+
						"**Step 7:** Availability preferences\n"+
						"**Step 8:** Document upload (optional)\n"+
						"**Step 9:** Final review & confirmation\n\n"+
						"⏱️ **Estimated time:** 5-10 minutes\n"+
						"💾 **Progress is saved automatically**",
					markup,
				)
			}).
				AssignCallback("back_to_welcome", "start_from_info").
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					if c.Callback() == nil {
						return nil
					}

					switch c.Callback().Unique {
					case "back_to_welcome":
						return ctrl.Jump("welcome")
					case "start_from_info":
						return ctrl.Jump("basic_info")
					}
					return nil
				}),

			// 3. Basic Information Collection
			intele.NewStep("basic_info", func(c tele.Context, ctrl intele.Controller) error {
				collector := ctrl.MessageCollector()

				markup := &tele.ReplyMarkup{}
				backBtn := markup.Data("⬅️ Back", "back_to_info_detail")
				markup.Inline(markup.Row(backBtn))

				return collector.Send(
					"📝 **Step 1/9: Basic Information**\n\n"+
						"Let's start with your full name.\n\n"+
						"Please enter your **first and last name**:",
					markup,
				)
			}).
				AssignCallback("back_to_info_detail").
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					if c.Callback() != nil && c.Callback().Unique == "back_to_info_detail" {
						return ctrl.Jump("info_detail")
					}

					if c.Text() == "" {
						return nil
					}

					collector := ctrl.MessageCollector()
					if c.Message() != nil {
						_ = collector.Collect(c.Message().ID, c.Message().Chat.ID)
					}

					validator, _ := intele.GetTyped[*ValidationService](ctrl.Container(), "validator")
					name := strings.TrimSpace(c.Text())
					if err := validator.ValidateName(name); err != nil {
						return collector.Send(fmt.Sprintf("❌ %s\n\nPlease try again:", err.Error()))
					}

					ctrl.Storage().Set("full_name", name)
					return ctrl.Jump("age_input")
				}).
				RequireDeps("validator"),

			// 4. Age Input
			intele.NewStep("age_input", func(c tele.Context, ctrl intele.Controller) error {
				collector := ctrl.MessageCollector()
				name, _ := ctrl.Storage().GetString("full_name")

				markup := &tele.ReplyMarkup{}
				backBtn := markup.Data("⬅️ Back", "back_to_basic_info")
				markup.Inline(markup.Row(backBtn))

				return collector.Send(fmt.Sprintf(
					"🎂 **Nice to meet you, %s!**\n\n"+
						"How old are you?\n\n"+
						"Please enter your age (number between 16-100):",
					name,
				), markup)
			}).
				AssignCallback("back_to_basic_info").
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					if c.Callback() != nil && c.Callback().Unique == "back_to_basic_info" {
						return ctrl.Jump("basic_info")
					}

					if c.Text() == "" {
						return nil
					}

					collector := ctrl.MessageCollector()
					if c.Message() != nil {
						_ = collector.Collect(c.Message().ID, c.Message().Chat.ID)
					}

					ageStr := strings.TrimSpace(c.Text())
					age, err := strconv.Atoi(ageStr)
					if err != nil || age < 16 || age > 100 {
						return collector.Send("❌ Please enter a valid age between 16 and 100:")
					}

					ctrl.Storage().Set("age", age)
					return ctrl.Jump("email_input")
				}),

			// 5. Email Input
			intele.NewStep("email_input", func(c tele.Context, ctrl intele.Controller) error {
				collector := ctrl.MessageCollector()

				markup := &tele.ReplyMarkup{}
				backBtn := markup.Data("⬅️ Back", "back_to_age_input")
				markup.Inline(markup.Row(backBtn))

				return collector.Send(
					"📧 **Email Address**\n\n"+
						"Please provide your email address.\n\n"+
						"This will be used for account identification.",
					markup,
				)
			}).
				AssignCallback("back_to_age_input").
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					if c.Callback() != nil && c.Callback().Unique == "back_to_age_input" {
						return ctrl.Jump("age_input")
					}

					if c.Text() == "" {
						return nil
					}

					collector := ctrl.MessageCollector()
					if c.Message() != nil {
						_ = collector.Collect(c.Message().ID, c.Message().Chat.ID)
					}

					validator, _ := intele.GetTyped[*ValidationService](ctrl.Container(), "validator")
					email := strings.TrimSpace(c.Text())
					if err := validator.ValidateEmail(email); err != nil {
						return collector.Send(fmt.Sprintf("❌ %s\n\nPlease try again:", err.Error()))
					}

					ctrl.Storage().Set("email", email)
					return ctrl.Jump("phone_contact")
				}).
				RequireDeps("validator"),

			// 6. Phone Contact Collection
			intele.NewStep("phone_contact", func(c tele.Context, ctrl intele.Controller) error {
				collector := ctrl.MessageCollector()

				shareBtn := &tele.ReplyButton{Contact: true, Text: "📱 Share My Contact"}
				skipBtn := &tele.ReplyButton{Text: "⏭️ Skip Phone Number"}
				backBtn := &tele.ReplyButton{Text: "⬅️ Back to Email"}

				return collector.Send(
					"📞 **Step 2/9: Contact Information**\n\n"+
						"Would you like to share your phone number?\n\n"+
						"You can either:\n"+
						"• Tap 'Share My Contact' to share automatically\n"+
						"• Type your phone number manually\n"+
						"• Skip this step\n\n"+
						"Your phone number helps with account security.",
					&tele.ReplyMarkup{
						OneTimeKeyboard: true,
						ReplyKeyboard: [][]tele.ReplyButton{
							{*shareBtn},
							{*skipBtn},
							{*backBtn},
						},
					},
				)
			}).
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					collector := ctrl.MessageCollector()

					if c.Message() != nil && c.Message().Contact != nil {
						_ = collector.Collect(c.Message().ID, c.Message().Chat.ID)
						contact := c.Message().Contact
						ctrl.Storage().Set("phone", contact.PhoneNumber)
						ctrl.Storage().Set("contact_shared", true)

						_ = collector.Send(
							"✅ **Contact received!**\n\n"+
								fmt.Sprintf("📱 Phone: %s", contact.PhoneNumber),
							&tele.ReplyMarkup{RemoveKeyboard: true},
						)
						return ctrl.Jump("location_request")
					}

					if c.Text() == "⏭️ Skip Phone Number" {
						_ = collector.Collect(c.Message().ID, c.Message().Chat.ID)
						ctrl.Storage().Set("phone", "")
						ctrl.Storage().Set("contact_shared", false)
						_ = collector.Send("⏭️ **Phone number skipped**", &tele.ReplyMarkup{RemoveKeyboard: true})
						return ctrl.Jump("location_request")
					}

					if c.Text() == "⬅️ Back to Email" {
						_ = collector.Collect(c.Message().ID, c.Message().Chat.ID)
						_ = collector.Send("⬅️ **Going back to email step**", &tele.ReplyMarkup{RemoveKeyboard: true})
						return ctrl.Jump("email_input")
					}

					if c.Text() != "" {
						_ = collector.Collect(c.Message().ID, c.Message().Chat.ID)
						phone := strings.TrimSpace(c.Text())

						if len(phone) < 10 {
							return collector.Send("❌ Please enter a valid phone number (at least 10 digits):")
						}

						ctrl.Storage().Set("phone", phone)
						ctrl.Storage().Set("contact_shared", false)
						_ = collector.Send("✅ **Phone number saved!**", &tele.ReplyMarkup{RemoveKeyboard: true})
						return ctrl.Jump("location_request")
					}

					return nil
				}),

			// 7. Location Request
			intele.NewStep("location_request", func(c tele.Context, ctrl intele.Controller) error {
				collector := ctrl.MessageCollector()

				shareLocationBtn := &tele.ReplyButton{Location: true, Text: "📍 Share My Location"}
				manualBtn := &tele.ReplyButton{Text: "✏️ Enter Manually"}
				skipBtn := &tele.ReplyButton{Text: "⏭️ Skip Location"}
				backBtn := &tele.ReplyButton{Text: "⬅️ Back to Contact"}

				return collector.Send(
					"📍 **Step 3/9: Location Information**\n\n"+
						"Where are you located?\n\n"+
						"You can:\n"+
						"• Share your current location\n"+
						"• Enter your city/country manually\n"+
						"• Skip this step\n\n"+
						"Location helps us provide relevant content.",
					&tele.ReplyMarkup{
						OneTimeKeyboard: true,
						ReplyKeyboard: [][]tele.ReplyButton{
							{*shareLocationBtn},
							{*manualBtn},
							{*skipBtn},
							{*backBtn},
						},
					},
				)
			}).
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					collector := ctrl.MessageCollector()

					if c.Message() != nil && c.Message().Location != nil {
						_ = collector.Collect(c.Message().ID, c.Message().Chat.ID)
						location := c.Message().Location
						ctrl.Storage().Set("latitude", location.Lat)
						ctrl.Storage().Set("longitude", location.Lng)
						ctrl.Storage().Set("location_shared", true)

						_ = collector.Send(
							fmt.Sprintf("✅ **Location received!**\n\n📍 Coordinates: %.4f, %.4f",
								location.Lat, location.Lng),
							&tele.ReplyMarkup{RemoveKeyboard: true},
						)
						return ctrl.Jump("profession_info")
					}

					if c.Text() != "" {
						_ = collector.Collect(c.Message().ID, c.Message().Chat.ID)

						switch c.Text() {
						case "✏️ Enter Manually":
							return ctrl.Jump("manual_location_input")
						case "⏭️ Skip Location":
							ctrl.Storage().Set("location", "Not provided")
							_ = collector.Send("⏭️ **Location skipped**", &tele.ReplyMarkup{RemoveKeyboard: true})
							return ctrl.Jump("profession_info")
						case "⬅️ Back to Contact":
							_ = collector.Send("⬅️ **Going back to contact step**", &tele.ReplyMarkup{RemoveKeyboard: true})
							return ctrl.Jump("phone_contact")
						}
					}

					return nil
				}),

			// 8. Manual Location Input
			intele.NewStep("manual_location_input", func(c tele.Context, ctrl intele.Controller) error {
				collector := ctrl.MessageCollector()

				markup := &tele.ReplyMarkup{}
				backBtn := markup.Data("⬅️ Back", "back_to_location_request")
				markup.Inline(markup.Row(backBtn))

				return collector.Send(
					"🌍 **Enter Your Location**\n\n"+
						"Please enter your city and country.\n\n"+
						"Example: `New York, USA` or `London, UK`",
					&tele.ReplyMarkup{RemoveKeyboard: true},
					markup,
				)
			}).
				AssignCallback("back_to_location_request").
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					if c.Callback() != nil && c.Callback().Unique == "back_to_location_request" {
						return ctrl.Jump("location_request")
					}

					if c.Text() == "" {
						return nil
					}

					collector := ctrl.MessageCollector()
					if c.Message() != nil {
						_ = collector.Collect(c.Message().ID, c.Message().Chat.ID)
					}

					location := strings.TrimSpace(c.Text())
					if len(location) < 3 {
						return collector.Send("❌ Please enter a valid location (at least 3 characters):")
					}

					ctrl.Storage().Set("location", location)
					ctrl.Storage().Set("location_shared", false)

					_ = collector.Send("✅ **Location saved!**")
					return ctrl.Jump("profession_info")
				}),

			// 9. Professional Information
			intele.NewStep("profession_info", func(c tele.Context, ctrl intele.Controller) error {
				collector := ctrl.MessageCollector()

				markup := &tele.ReplyMarkup{}
				student := markup.Data("🎓 Student", "prof_student")
				employed := markup.Data("💼 Employed", "prof_employed")
				freelancer := markup.Data("💻 Freelancer", "prof_freelancer")
				entrepreneur := markup.Data("🚀 Entrepreneur", "prof_entrepreneur")
				unemployed := markup.Data("🔍 Looking for work", "prof_unemployed")
				retired := markup.Data("🏖️ Retired", "prof_retired")
				other := markup.Data("📝 Other", "prof_other")
				backBtn := markup.Data("⬅️ Back", "back_to_location")

				markup.Inline(
					markup.Row(student, employed),
					markup.Row(freelancer, entrepreneur),
					markup.Row(unemployed, retired),
					markup.Row(other),
					markup.Row(backBtn),
				)

				return collector.Send(
					"💼 **Step 4/9: Professional Background**\n\n"+
						"What best describes your current professional status?",
					markup,
				)
			}).
				AssignCallback("prof_student", "prof_employed", "prof_freelancer", "prof_entrepreneur", "prof_unemployed", "prof_retired", "prof_other", "back_to_location").
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					if c.Callback() == nil {
						return nil
					}

					if c.Callback().Unique == "back_to_location" {
						if ctrl.Storage().Has("location_shared") {
							return ctrl.Jump("location_request")
						} else {
							return ctrl.Jump("manual_location_input")
						}
					}

					profession := strings.TrimPrefix(c.Callback().Unique, "prof_")
					ctrl.Storage().Set("profession", profession)

					if profession == "other" {
						return ctrl.Jump("custom_profession")
					}

					return ctrl.Jump("skills_interests")
				}),

			// 10. Custom Profession Input
			intele.NewStep("custom_profession", func(c tele.Context, ctrl intele.Controller) error {
				collector := ctrl.MessageCollector()

				markup := &tele.ReplyMarkup{}
				backBtn := markup.Data("⬅️ Back", "back_to_profession_info")
				markup.Inline(markup.Row(backBtn))

				return collector.Send(
					"📝 **Custom Profession**\n\n"+
						"Please describe your professional status or occupation:",
					markup,
				)
			}).
				AssignCallback("back_to_profession_info").
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					if c.Callback() != nil && c.Callback().Unique == "back_to_profession_info" {
						return ctrl.Jump("profession_info")
					}

					if c.Text() == "" {
						return nil
					}

					collector := ctrl.MessageCollector()
					if c.Message() != nil {
						_ = collector.Collect(c.Message().ID, c.Message().Chat.ID)
					}

					profession := strings.TrimSpace(c.Text())
					if len(profession) < 2 {
						return collector.Send("❌ Please provide a more detailed description:")
					}

					ctrl.Storage().Set("custom_profession", profession)
					return ctrl.Jump("skills_interests")
				}),

			// 11. Skills and Interests Selection
			intele.NewStep("skills_interests", func(c tele.Context, ctrl intele.Controller) error {
				selectedSkills, _ := ctrl.Storage().Get("selected_skills")
				if selectedSkills == nil {
					selectedSkills = make([]string, 0)
				}

				skills := selectedSkills.([]string)
				markup := &tele.ReplyMarkup{}
				tech := markup.Data("💻 Technology", "skill_tech")
				design := markup.Data("🎨 Design", "skill_design")
				business := markup.Data("📊 Business", "skill_business")
				marketing := markup.Data("📢 Marketing", "skill_marketing")
				education := markup.Data("📚 Education", "skill_education")
				healthcare := markup.Data("🏥 Healthcare", "skill_healthcare")
				arts := markup.Data("🎭 Arts & Culture", "skill_arts")
				sports := markup.Data("⚽ Sports", "skill_sports")
				done := markup.Data("✅ Done Selecting", "skills_done")
				backBtn := markup.Data("⬅️ Back", "back_to_profession")

				for i, btn := range []*tele.Btn{&tech, &design, &business, &marketing, &education, &healthcare, &arts, &sports} {
					skillName := getSkillName([]string{"tech", "design", "business", "marketing", "education", "healthcare", "arts", "sports"}[i])
					for _, selected := range skills {
						if selected == skillName {
							btn.Text = "✅ " + btn.Text
							break
						}
					}
				}

				markup.Inline(
					markup.Row(tech, design),
					markup.Row(business, marketing),
					markup.Row(education, healthcare),
					markup.Row(arts, sports),
					markup.Row(done),
					markup.Row(backBtn),
				)

				var skillsList string
				if len(skills) == 0 {
					skillsList = "_None selected yet_"
				} else {
					skillsList = strings.Join(skills, ", ")
				}

				text := "🎯 **Step 5/9: Skills & Interests**\n\n" +
					"Select your areas of interest or expertise.\n" +
					"You can select multiple options.\n\n" +
					"**Selected:** " + skillsList + "\n\n" +
					"Click 'Done Selecting' when finished."

				if c.Callback() != nil {
					return c.Edit(text, markup)
				} else {
					collector := ctrl.MessageCollector()
					return collector.Send(text, markup)
				}
			}).
				AssignCallback("skill_tech", "skill_design", "skill_business", "skill_marketing",
					"skill_education", "skill_healthcare", "skill_arts", "skill_sports", "skills_done", "back_to_profession").
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					if c.Callback() == nil {
						return nil
					}

					if c.Callback().Unique == "back_to_profession" {
						if ctrl.Storage().Has("custom_profession") {
							return ctrl.Jump("custom_profession")
						} else {
							return ctrl.Jump("profession_info")
						}
					}

					if c.Callback().Unique == "skills_done" {
						selectedSkills, exists := ctrl.Storage().Get("selected_skills")
						if !exists || len(selectedSkills.([]string)) == 0 {
							return c.Edit("❌ Please select at least one skill or interest area.")
						}
						return ctrl.Jump("experience_rating")
					}

					skill := strings.TrimPrefix(c.Callback().Unique, "skill_")
					selectedSkills, exists := ctrl.Storage().Get("selected_skills")
					if !exists {
						selectedSkills = make([]string, 0)
					}

					skills := selectedSkills.([]string)
					skillName := getSkillName(skill)
					found := false

					for i, s := range skills {
						if s == skillName {
							skills = append(skills[:i], skills[i+1:]...)
							found = true
							break
						}
					}

					if !found {
						skills = append(skills, skillName)
					}

					ctrl.Storage().Set("selected_skills", skills)
					return ctrl.Jump("skills_interests")
				}),

			// 12. Experience Rating
			intele.NewStep("experience_rating", func(c tele.Context, ctrl intele.Controller) error {
				collector := ctrl.MessageCollector()

				markup := &tele.ReplyMarkup{}
				rating1 := markup.Data("⭐", "rating_1")
				rating2 := markup.Data("⭐⭐", "rating_2")
				rating3 := markup.Data("⭐⭐⭐", "rating_3")
				rating4 := markup.Data("⭐⭐⭐⭐", "rating_4")
				rating5 := markup.Data("⭐⭐⭐⭐⭐", "rating_5")
				backBtn := markup.Data("⬅️ Back", "back_to_skills")

				markup.Inline(
					markup.Row(rating1),
					markup.Row(rating2),
					markup.Row(rating3),
					markup.Row(rating4),
					markup.Row(rating5),
					markup.Row(backBtn),
				)

				return collector.Send(
					"⭐ **Step 6/9: Experience Rating**\n\n"+
						"How would you rate your overall professional experience?\n\n"+
						"⭐ = Beginner\n"+
						"⭐⭐ = Some experience\n"+
						"⭐⭐⭐ = Intermediate\n"+
						"⭐⭐⭐⭐ = Advanced\n"+
						"⭐⭐⭐⭐⭐ = Expert",
					markup,
				)
			}).
				AssignCallback("rating_1", "rating_2", "rating_3", "rating_4", "rating_5", "back_to_skills").
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					if c.Callback() == nil {
						return nil
					}

					if c.Callback().Unique == "back_to_skills" {
						return ctrl.Jump("skills_interests")
					}

					rating := strings.TrimPrefix(c.Callback().Unique, "rating_")
					ratingNum, _ := strconv.Atoi(rating)
					ctrl.Storage().Set("experience_rating", ratingNum)

					return ctrl.Jump("availability")
				}),

			// 13. Availability Preferences
			intele.NewStep("availability", func(c tele.Context, ctrl intele.Controller) error {
				collector := ctrl.MessageCollector()

				markup := &tele.ReplyMarkup{}
				fulltime := markup.Data("🕘 Full-time", "avail_fulltime")
				parttime := markup.Data("🕐 Part-time", "avail_parttime")
				flexible := markup.Data("🕑 Flexible", "avail_flexible")
				weekends := markup.Data("🏖️ Weekends only", "avail_weekends")
				evenings := markup.Data("🌆 Evenings", "avail_evenings")
				backBtn := markup.Data("⬅️ Back", "back_to_experience")

				markup.Inline(
					markup.Row(fulltime, parttime),
					markup.Row(flexible, weekends),
					markup.Row(evenings),
					markup.Row(backBtn),
				)

				return collector.Send(
					"📅 **Step 7/9: Availability Preferences**\n\n"+
						"What's your preferred availability for projects or opportunities?",
					markup,
				)
			}).
				AssignCallback("avail_fulltime", "avail_parttime", "avail_flexible", "avail_weekends", "avail_evenings", "back_to_experience").
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					if c.Callback() == nil {
						return nil
					}

					if c.Callback().Unique == "back_to_experience" {
						return ctrl.Jump("experience_rating")
					}

					availability := strings.TrimPrefix(c.Callback().Unique, "avail_")
					ctrl.Storage().Set("availability", availability)

					return ctrl.Jump("document_upload")
				}),

			// 14. Document Upload (Optional)
			intele.NewStep("document_upload", func(c tele.Context, ctrl intele.Controller) error {
				collector := ctrl.MessageCollector()

				markup := &tele.ReplyMarkup{}
				skipBtn := markup.Data("⏭️ Skip Upload", "skip_upload")
				backBtn := markup.Data("⬅️ Back", "back_to_availability")

				markup.Inline(
					markup.Row(skipBtn),
					markup.Row(backBtn),
				)

				return collector.Send(
					"📎 **Step 8/9: Document Upload (Optional)**\n\n"+
						"You can upload a document such as:\n"+
						"• Resume/CV\n"+
						"• Portfolio\n"+
						"• Certificate\n"+
						"• Any relevant document\n\n"+
						"Simply send any document or photo, or skip this step.\n\n"+
						"**Supported formats:** PDF, DOC, JPG, PNG",
					markup,
				)
			}).
				AssignCallback("skip_upload", "back_to_availability").
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					collector := ctrl.MessageCollector()

					if c.Callback() != nil && c.Callback().Unique == "skip_upload" {
						ctrl.Storage().Set("document_uploaded", false)
						return ctrl.Jump("final_review")
					}

					if c.Callback() != nil && c.Callback().Unique == "back_to_availability" {
						return ctrl.Jump("availability")
					}

					if c.Message() != nil && (c.Message().Document != nil || c.Message().Photo != nil) {
						_ = collector.Collect(c.Message().ID, c.Message().Chat.ID)

						var fileName string
						var fileSize int64

						if c.Message().Document != nil {
							fileName = c.Message().Document.FileName
							fileSize = c.Message().Document.FileSize
						} else if c.Message().Photo != nil {
							fileName = "photo.jpg"
							fileSize = int64(c.Message().Photo.FileSize)
						}

						ctrl.Storage().Set("document_uploaded", true)
						ctrl.Storage().Set("document_name", fileName)
						ctrl.Storage().Set("document_size", fileSize)

						_ = collector.Send(
							fmt.Sprintf("✅ **Document received!**\n\n📄 **File:** %s\n📏 **Size:** %.1f KB",
								fileName, float64(fileSize)/1024),
						)
						return ctrl.Jump("final_review")
					}

					return nil
				}),

			// 15. Final Review and Confirmation
			intele.NewStep("final_review", func(c tele.Context, ctrl intele.Controller) error {
				collector := ctrl.MessageCollector()

				name, _ := ctrl.Storage().GetString("full_name")
				age, _ := ctrl.Storage().GetInt("age")
				email, _ := ctrl.Storage().GetString("email")
				phone, _ := ctrl.Storage().GetString("phone")
				location, _ := ctrl.Storage().GetString("location")
				profession, _ := ctrl.Storage().GetString("profession")
				customProfession, _ := ctrl.Storage().GetString("custom_profession")
				skills, _ := ctrl.Storage().Get("selected_skills")
				rating, _ := ctrl.Storage().GetInt("experience_rating")
				availability, _ := ctrl.Storage().GetString("availability")
				docUploaded, _ := ctrl.Storage().GetBool("document_uploaded")

				var skillsStr string
				if skills != nil {
					skillsStr = strings.Join(skills.([]string), ", ")
				}

				var professionStr string
				if customProfession != "" {
					professionStr = customProfession
				} else {
					professionStr = profession
				}

				var locationStr string
				if location == "" || location == "Not provided" {
					locationStr = "Not provided"
				} else {
					locationStr = location
				}

				var phoneStr string
				if phone == "" {
					phoneStr = "Not provided"
				} else {
					phoneStr = phone
				}

				var docStr string
				if docUploaded {
					docName, _ := ctrl.Storage().GetString("document_name")
					docStr = fmt.Sprintf("✅ %s", docName)
				} else {
					docStr = "❌ No document uploaded"
				}

				markup := &tele.ReplyMarkup{}
				confirmBtn := markup.Data("✅ Confirm & Submit", "confirm_registration")
				editBtn := markup.Data("✏️ Edit Information", "edit_info")
				cancelBtn := markup.Data("❌ Cancel Registration", "cancel_registration")

				markup.Inline(
					markup.Row(confirmBtn),
					markup.Row(editBtn),
					markup.Row(cancelBtn),
				)

				return collector.Send(
					fmt.Sprintf(
						"📋 **Step 9/9: Final Review**\n\n"+
							"Please review your information:\n\n"+
							"👤 **Name:** %s\n"+
							"🎂 **Age:** %d\n"+
							"📧 **Email:** %s\n"+
							"📱 **Phone:** %s\n"+
							"📍 **Location:** %s\n"+
							"💼 **Profession:** %s\n"+
							"🎯 **Skills:** %s\n"+
							"⭐ **Experience:** %s\n"+
							"📅 **Availability:** %s\n"+
							"📎 **Document:** %s\n\n"+
							"Is everything correct?",
						name, age, email, phoneStr, locationStr, professionStr, skillsStr,
						getExperienceText(rating), availability, docStr,
					),
					markup,
				)
			}).
				AssignCallback("confirm_registration", "edit_info", "cancel_registration").
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					if c.Callback() == nil {
						return nil
					}

					switch c.Callback().Unique {
					case "confirm_registration":
						return ctrl.Jump("save_user")
					case "edit_info":
						return ctrl.Jump("edit_selection")
					case "cancel_registration":
						return ctrl.Cancel()
					}
					return nil
				}),

			// 16. Edit Selection
			intele.NewStep("edit_selection", func(c tele.Context, ctrl intele.Controller) error {
				collector := ctrl.MessageCollector()

				markup := &tele.ReplyMarkup{}
				editBasic := markup.Data("📝 Basic Info", "edit_basic")
				editContact := markup.Data("📞 Contact", "edit_contact")
				editLocation := markup.Data("📍 Location", "edit_location")
				editProfession := markup.Data("💼 Profession", "edit_profession")
				editSkills := markup.Data("🎯 Skills", "edit_skills")
				editRating := markup.Data("⭐ Experience", "edit_rating")
				editAvailability := markup.Data("📅 Availability", "edit_availability")
				editDocument := markup.Data("📎 Document", "edit_document")
				backBtn := markup.Data("⬅️ Back to Review", "back_to_review")

				markup.Inline(
					markup.Row(editBasic, editContact),
					markup.Row(editLocation, editProfession),
					markup.Row(editSkills, editRating),
					markup.Row(editAvailability, editDocument),
					markup.Row(backBtn),
				)

				return collector.Send(
					"✏️ **Edit Information**\n\n"+
						"Which section would you like to modify?",
					markup,
				)
			}).
				AssignCallback("edit_basic", "edit_contact", "edit_location", "edit_profession",
					"edit_skills", "edit_rating", "edit_availability", "edit_document", "back_to_review").
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					if c.Callback() == nil {
						return nil
					}

					switch c.Callback().Unique {
					case "edit_basic":
						return ctrl.Jump("basic_info")
					case "edit_contact":
						return ctrl.Jump("phone_contact")
					case "edit_location":
						return ctrl.Jump("location_request")
					case "edit_profession":
						return ctrl.Jump("profession_info")
					case "edit_skills":
						ctrl.Storage().Delete("selected_skills")
						return ctrl.Jump("skills_interests")
					case "edit_rating":
						return ctrl.Jump("experience_rating")
					case "edit_availability":
						return ctrl.Jump("availability")
					case "edit_document":
						return ctrl.Jump("document_upload")
					case "back_to_review":
						return ctrl.Jump("final_review")
					}
					return nil
				}),

			// 17. Save User Data
			intele.NewStep("save_user", func(c tele.Context, ctrl intele.Controller) error {
				userService, _ := intele.GetTyped[*UserService](ctrl.Container(), "userService")
				logger, _ := intele.GetTyped[*Logger](ctrl.Container(), "logger")

				name, _ := ctrl.Storage().GetString("full_name")
				age, _ := ctrl.Storage().GetInt("age")
				email, _ := ctrl.Storage().GetString("email")
				phone, _ := ctrl.Storage().GetString("phone")
				location, _ := ctrl.Storage().GetString("location")
				profession, _ := ctrl.Storage().GetString("profession")
				customProfession, _ := ctrl.Storage().GetString("custom_profession")
				skills, _ := ctrl.Storage().Get("selected_skills")
				rating, _ := ctrl.Storage().GetInt("experience_rating")
				availability, _ := ctrl.Storage().GetString("availability")
				docUploaded, _ := ctrl.Storage().GetBool("document_uploaded")

				user := &User{
					TelegramID:       c.Sender().ID,
					FullName:         name,
					Age:              age,
					Email:            email,
					Phone:            phone,
					Location:         location,
					Profession:       profession,
					CustomProfession: customProfession,
					Skills:           skills.([]string),
					ExperienceRating: rating,
					Availability:     availability,
					DocumentUploaded: docUploaded,
					CreatedAt:        time.Now(),
				}

				if err := userService.SaveUser(user); err != nil {
					logger.Error("Failed to save user registration", err)
					collector := ctrl.MessageCollector()
					return collector.Send("❌ **Error saving registration.**\n\nPlease try again later.")
				}

				logger.Info("User registration completed successfully", user.FullName, user.Email)

				collector := ctrl.MessageCollector()
				_ = collector.Clear(intele.ClearOptions{
					IgnoreErrors: true,
					ExcludeLast:  true,
				})

				_ = ctrl.Complete()
				return c.Send(
					fmt.Sprintf(
						"🎉 **Registration Complete!**\n\n"+
							"Welcome, **%s**!\n\n"+
							"Your comprehensive profile has been created successfully.\n\n"+
							"✅ **Registration Summary:**\n"+
							"• Your information is stored\n"+
							"• Profile is ready to use\n"+
							"• You can view it anytime\n\n"+
							"Thank you for completing the registration process!",
						name,
					),
				)
			}).
				RequireDeps("userService", "logger"),
		).
		StartWith("welcome").
		WithTimeout(30 * time.Minute).
		WithContainer(container).
		WithMiddleware(intele.LoggingMiddleware(func(format string, args ...interface{}) {
			log.Printf("[REGISTRATION] "+format, args...)
		})).
		OnComplete(func(c tele.Context, ctrl intele.Controller) error {
			collector := ctrl.MessageCollector()
			_ = collector.Clear(intele.ClearOptions{IgnoreErrors: true})
			return nil
		}).
		OnError(func(c tele.Context, ctrl intele.Controller, err error) error {
			log.Printf("Registration error: %v", err)
			return c.Send("❌ **Registration error occurred.**\n\nPlease try again with /register")
		}).
		OnCancel(func(c tele.Context, ctrl intele.Controller) error {
			collector := ctrl.MessageCollector()
			_ = collector.Clear(intele.ClearOptions{IgnoreErrors: true})
			return c.Send("🚫 **Registration cancelled.**\n\nAll your data has been cleared.")
		}).
		OnTimeout(func(c tele.Context, ctrl intele.Controller) error {
			collector := ctrl.MessageCollector()
			_ = collector.Clear(intele.ClearOptions{IgnoreErrors: true})
			return c.Send("⏰ **Registration timed out.**\n\nPlease start over with /register")
		}).
		Build()

	if err != nil {
		log.Fatal("Failed to create registration flow:", err)
	}

	return flow
}

// Helper functions
func getProgressText(step string) string {
	steps := map[string]string{
		"welcome":               "0% - Getting started",
		"info_detail":           "5% - Reading information",
		"basic_info":            "15% - Basic information",
		"age_input":             "20% - Age verification",
		"email_input":           "25% - Email setup",
		"phone_contact":         "35% - Contact information",
		"location_request":      "45% - Location setup",
		"manual_location_input": "50% - Location details",
		"profession_info":       "60% - Professional background",
		"custom_profession":     "65% - Custom profession",
		"skills_interests":      "75% - Skills & interests",
		"experience_rating":     "80% - Experience rating",
		"availability":          "85% - Availability preferences",
		"document_upload":       "90% - Document upload",
		"final_review":          "95% - Final review",
		"edit_selection":        "95% - Editing information",
		"save_user":             "100% - Saving registration",
	}

	if progress, exists := steps[step]; exists {
		return progress
	}
	return "Unknown step"
}

func getSkillName(skill string) string {
	skillNames := map[string]string{
		"tech":       "Technology",
		"design":     "Design",
		"business":   "Business",
		"marketing":  "Marketing",
		"education":  "Education",
		"healthcare": "Healthcare",
		"arts":       "Arts & Culture",
		"sports":     "Sports",
	}

	if name, exists := skillNames[skill]; exists {
		return name
	}
	return skill
}

func getExperienceText(rating int) string {
	switch rating {
	case 1:
		return "⭐ Beginner"
	case 2:
		return "⭐⭐ Some experience"
	case 3:
		return "⭐⭐⭐ Intermediate"
	case 4:
		return "⭐⭐⭐⭐ Advanced"
	case 5:
		return "⭐⭐⭐⭐⭐ Expert"
	default:
		return "Not rated"
	}
}

type User struct {
	TelegramID       int64     `json:"telegram_id"`
	FullName         string    `json:"full_name"`
	Age              int       `json:"age"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	Location         string    `json:"location"`
	Profession       string    `json:"profession"`
	CustomProfession string    `json:"custom_profession"`
	Skills           []string  `json:"skills"`
	ExperienceRating int       `json:"experience_rating"`
	Availability     string    `json:"availability"`
	DocumentUploaded bool      `json:"document_uploaded"`
	CreatedAt        time.Time `json:"created_at"`
}

type UserService struct {
	users map[int64]*User
}

func (s *UserService) SaveUser(user *User) error {
	if s.users == nil {
		s.users = make(map[int64]*User)
	}
	s.users[user.TelegramID] = user
	log.Printf("User registered: %s (%s)", user.FullName, user.Email)
	return nil
}

func (s *UserService) GetUser(telegramID int64) (*User, error) {
	if s.users == nil {
		s.users = make(map[int64]*User)
	}

	user, exists := s.users[telegramID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

type ValidationService struct{}

func (v *ValidationService) ValidateName(name string) error {
	if len(name) < 2 {
		return fmt.Errorf("name must be at least 2 characters long")
	}
	if len(name) > 100 {
		return fmt.Errorf("name must be less than 100 characters")
	}
	if !strings.Contains(name, " ") {
		return fmt.Errorf("please enter both first and last name")
	}
	return nil
}

func (v *ValidationService) ValidateEmail(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return fmt.Errorf("invalid email format")
	}
	if len(email) < 5 {
		return fmt.Errorf("email too short")
	}
	if len(email) > 100 {
		return fmt.Errorf("email too long")
	}
	return nil
}

type Logger struct{}

func (l *Logger) Info(message string, args ...interface{}) {
	log.Printf("[INFO] %s %v", message, args)
}

func (l *Logger) Error(message string, err error) {
	log.Printf("[ERROR] %s: %v", message, err)
}
