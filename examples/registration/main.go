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

// User represents a registered user
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

// UserService interface for user operations
type UserService interface {
	SaveUser(user *User) error
	GetUser(telegramID int64) (*User, error)
	GetAllUsers() ([]*User, error)
}

// ValidationService interface for validation operations
type ValidationService interface {
	ValidateName(name string) error
	ValidateEmail(email string) error
	ValidatePhone(phone string) error
}

// Logger interface for logging operations
type Logger interface {
	Info(message string, args ...interface{})
	Error(message string, args ...interface{})
	Debug(message string, args ...interface{})
}

// userService implements UserService interface
type userService struct {
	users map[int64]*User
}

func (us *userService) SaveUser(user *User) error {
	us.users[user.TelegramID] = user
	return nil
}

func (us *userService) GetUser(telegramID int64) (*User, error) {
	user, exists := us.users[telegramID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (us *userService) GetAllUsers() ([]*User, error) {
	users := make([]*User, 0, len(us.users))
	for _, user := range us.users {
		users = append(users, user)
	}
	return users, nil
}

// validationService implements ValidationService interface
type validationService struct{}

func (vs *validationService) ValidateName(name string) error {
	if len(name) < 2 {
		return fmt.Errorf("name must be at least 2 characters long")
	}
	if len(name) > 50 {
		return fmt.Errorf("name must be less than 50 characters")
	}
	return nil
}

func (vs *validationService) ValidateEmail(email string) error {
	if !strings.Contains(email, "@") {
		return fmt.Errorf("please enter a valid email address")
	}
	if len(email) < 5 {
		return fmt.Errorf("email address is too short")
	}
	return nil
}

func (vs *validationService) ValidatePhone(phone string) error {
	if len(phone) < 10 {
		return fmt.Errorf("phone number must be at least 10 digits")
	}
	return nil
}

// logger implements Logger interface
type logger struct{}

func (l *logger) Info(message string, args ...interface{}) {
	log.Printf("[INFO] "+message, args...)
}

func (l *logger) Error(message string, args ...interface{}) {
	log.Printf("[ERROR] "+message, args...)
}

func (l *logger) Debug(message string, args ...interface{}) {
	log.Printf("[DEBUG] "+message, args...)
}

func main() {
	// Initialize Telegram bot
	bot, err := tele.NewBot(tele.Settings{
		Token:     "YOUR_BOT_TOKEN",
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
				"• Dependency injection with interface validation\n" +
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
	container.Set("userService", &userService{users: make(map[int64]*User)})
	container.Set("validator", &validationService{})
	container.Set("logger", &logger{})
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
				RequireDeps(
					intele.NewDeps().
						Require("validator", (*ValidationService)(nil)),
				).
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

					validator, _ := intele.GetTyped[ValidationService](ctrl.Container(), "validator")
					name := strings.TrimSpace(c.Text())
					if err := validator.ValidateName(name); err != nil {
						return collector.Send(fmt.Sprintf("❌ %s\n\nPlease try again:", err.Error()))
					}

					ctrl.Storage().Set("full_name", name)
					return ctrl.Jump("age_input")
				}),

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
				RequireDeps(
					intele.NewDeps().
						Require("validator", (*ValidationService)(nil)),
				).
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

					validator, _ := intele.GetTyped[ValidationService](ctrl.Container(), "validator")
					email := strings.TrimSpace(c.Text())
					if err := validator.ValidateEmail(email); err != nil {
						return collector.Send(fmt.Sprintf("❌ %s\n\nPlease try again:", err.Error()))
					}

					ctrl.Storage().Set("email", email)
					return ctrl.Jump("phone_contact")
				}),

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

			// 9. Save User Data
			intele.NewStep("save_user", func(c tele.Context, ctrl intele.Controller) error {
				userService, _ := intele.GetTyped[UserService](ctrl.Container(), "userService")
				logger, _ := intele.GetTyped[Logger](ctrl.Container(), "logger")

				name, _ := ctrl.Storage().GetString("full_name")
				age, _ := ctrl.Storage().GetInt("age")
				email, _ := ctrl.Storage().GetString("email")
				phone, _ := ctrl.Storage().GetString("phone")
				location, _ := ctrl.Storage().GetString("location")

				user := &User{
					TelegramID: c.Sender().ID,
					FullName:   name,
					Age:        age,
					Email:      email,
					Phone:      phone,
					Location:   location,
					CreatedAt:  time.Now(),
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
							"Your profile has been created successfully.\n\n"+
							"✅ **Registration Summary:**\n"+
							"• Your information is stored\n"+
							"• Profile is ready to use\n"+
							"• You can view it anytime\n\n"+
							"Thank you for completing the registration process!",
						name,
					),
				)
			}).
				RequireDeps(
					intele.NewDeps().
						Require("userService", (*UserService)(nil)).
						Require("logger", (*Logger)(nil)),
				),

			// Additional step for profession info
			intele.NewStep("profession_info", func(c tele.Context, ctrl intele.Controller) error {
				collector := ctrl.MessageCollector()

				markup := &tele.ReplyMarkup{}
				student := markup.Data("🎓 Student", "prof_student")
				employed := markup.Data("💼 Employed", "prof_employed")
				freelancer := markup.Data("💻 Freelancer", "prof_freelancer")
				other := markup.Data("📝 Other", "prof_other")
				skipBtn := markup.Data("⏭️ Skip", "prof_skip")

				markup.Inline(
					markup.Row(student, employed),
					markup.Row(freelancer, other),
					markup.Row(skipBtn),
				)

				return collector.Send(
					"💼 **Step 4/9: Professional Information**\n\n"+
						"What best describes your current status?",
					markup,
				)
			}).
				AssignCallback("prof_student", "prof_employed", "prof_freelancer", "prof_other", "prof_skip").
				OnComplete(func(c tele.Context, ctrl intele.Controller) error {
					if c.Callback() == nil {
						return nil
					}

					profession := strings.TrimPrefix(c.Callback().Unique, "prof_")
					if profession == "skip" {
						profession = "Not specified"
					}

					ctrl.Storage().Set("profession", profession)
					return ctrl.Jump("save_user")
				}),
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
			return c.Send("❌ **Registration cancelled.**")
		}).
		Build()

	if err != nil {
		log.Fatal("Failed to create registration flow:", err)
	}

	return flow
}

func getProgressText(currentStep string) string {
	switch currentStep {
	case "welcome":
		return "Getting started"
	case "info_detail":
		return "Reading information"
	case "basic_info":
		return "Collecting basic info (1/9)"
	case "age_input":
		return "Collecting basic info (1/9)"
	case "email_input":
		return "Collecting basic info (1/9)"
	case "phone_contact":
		return "Contact information (2/9)"
	case "location_request":
		return "Location information (3/9)"
	case "manual_location_input":
		return "Location information (3/9)"
	case "profession_info":
		return "Professional info (4/9)"
	case "save_user":
		return "Saving profile (9/9)"
	default:
		return "In progress"
	}
}
