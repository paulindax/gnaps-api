package config

import (
	"gnaps-api/controllers"
	"gnaps-api/repositories"
	"gnaps-api/services"
	"gnaps-api/workers"

	"gorm.io/gorm"
)

// PaymentWorker is exported for use in main.go
var PaymentWorker *workers.PaymentWorker

// MomoPaymentService is exported for use in worker
var MomoPaymentService *services.MomoPaymentService

// InitializeControllers sets up dependency injection for all refactored controllers
func InitializeControllers(db *gorm.DB) {
	// Initialize Repositories
	eventRepo := repositories.NewEventRepository(db)
	registrationRepo := repositories.NewRegistrationRepository(db)
	schoolRepo := repositories.NewSchoolRepository(db)
	newsRepo := repositories.NewNewsRepository(db)
	commentRepo := repositories.NewCommentRepository(db)
	userRepo := repositories.NewUserRepository(db)
	regionRepo := repositories.NewRegionRepository(db)
	zoneRepo := repositories.NewZoneRepository(db)
	groupRepo := repositories.NewGroupRepository(db)
	positionRepo := repositories.NewPositionRepository(db)
	executiveRepo := repositories.NewExecutiveRepository(db)
	contactPersonRepo := repositories.NewContactPersonRepository(db)
	documentRepo := repositories.NewDocumentRepository(db)
	financeAccountRepo := repositories.NewFinanceAccountRepository(db)
	billParticularRepo := repositories.NewBillParticularRepository(db)
	billRepo := repositories.NewBillRepository(db)
	billItemRepo := repositories.NewBillItemRepository(db)
	momoPaymentRepo := repositories.NewMomoPaymentRepository(db)
	activityLogRepo := repositories.NewActivityLogRepository(db)
	schoolBillRepo := repositories.NewSchoolBillRepository(db)

	// Initialize Services
	eventService := services.NewEventService(eventRepo, registrationRepo)
	schoolService := services.NewSchoolService(schoolRepo, userRepo)
	newsService := services.NewNewsService(newsRepo, commentRepo, userRepo)
	regionService := services.NewRegionService(regionRepo)
	zoneService := services.NewZoneService(zoneRepo)
	groupService := services.NewGroupService(groupRepo)
	positionService := services.NewPositionService(positionRepo)
	executiveService := services.NewExecutiveService(executiveRepo, userRepo)
	contactPersonService := services.NewContactPersonService(contactPersonRepo)
	documentService := services.NewDocumentService(documentRepo)
	dashboardService := services.NewDashboardService(db)
	mediaService := services.NewMediaService()
	financeAccountService := services.NewFinanceAccountService(financeAccountRepo)
	billParticularService := services.NewBillParticularService(billParticularRepo)
	billService := services.NewBillService(billRepo, billItemRepo)
	chatService := services.NewChatService()
	momoPaymentService := services.NewMomoPaymentService(momoPaymentRepo, registrationRepo, eventRepo, db)
	financeReportsService := services.NewFinanceReportsService(db)
	smsService := services.NewSmsService(db)
	activityLogService := services.NewActivityLogService(activityLogRepo)
	schoolBillService := services.NewSchoolBillService(schoolBillRepo)

	// Store globally for worker access
	MomoPaymentService = momoPaymentService

	// Initialize Payment Worker
	PaymentWorker = workers.NewPaymentWorker()

	// Start the payment processor (runs every 3 seconds to process "created" payments with Hubtel)
	PaymentWorker.StartPaymentProcessor(momoPaymentService.ProcessCreatedPayments)

	// Start the payment status checker (runs every 10 seconds to check "pending" payment statuses)
	PaymentWorker.StartStatusChecker(momoPaymentService.CheckAndUpdatePendingPayments)

	// Initialize Controllers
	publicEventsController := controllers.NewPublicEventsController(eventRepo, registrationRepo, schoolRepo, db)
	publicEventsController.SetPaymentDependencies(momoPaymentService, PaymentWorker)
	publicController := controllers.NewPublicController(regionRepo, zoneRepo, schoolRepo, contactPersonRepo, db)
	paymentsController := controllers.NewPaymentsController(momoPaymentService, PaymentWorker)

	// Initialize Refactored Controllers
	eventsController := controllers.NewEventsController(eventService, schoolService)
	newsController := controllers.NewNewsController(newsService)
	schoolsController := controllers.NewSchoolsController(schoolService)
	regionsController := controllers.NewRegionsController(regionService)
	zonesController := controllers.NewZonesController(zoneService)
	groupsController := controllers.NewGroupsController(groupService)
	positionsController := controllers.NewPositionsController(positionService)
	executivesController := controllers.NewExecutivesController(executiveService)
	contactPersonsController := controllers.NewContactPersonsController(contactPersonService)
	documentsController := controllers.NewDocumentsController(documentService)
	dashboardController := controllers.NewDashboardController(dashboardService)
	mediaController := controllers.NewMediaController(mediaService)
	financeAccountsController := controllers.NewFinanceAccountsController(financeAccountService)
	billParticularsController := controllers.NewBillParticularsController(billParticularService)
	billsController := controllers.NewBillsController(billService)
	chatController := controllers.NewChatController(chatService)
	financeReportsController := controllers.NewFinanceReportsController(financeReportsService)
	smsController := controllers.NewSmsController(smsService, db)
	activityLogsController := controllers.NewActivityLogsController(activityLogService)
	schoolBillsController := controllers.NewSchoolBillsController(schoolBillService)
	schoolPaymentsController := controllers.NewSchoolPaymentsController(schoolBillService, momoPaymentService, PaymentWorker)

	// Register refactored controllers (these will override the old ones)
	controllers.RegisterController("events", eventsController)
	controllers.RegisterController("news", newsController)
	controllers.RegisterController("schools", schoolsController)
	controllers.RegisterController("regions", regionsController)
	controllers.RegisterController("zones", zonesController)
	controllers.RegisterController("groups", groupsController)
	controllers.RegisterController("positions", positionsController)
	controllers.RegisterController("executives", executivesController)
	controllers.RegisterController("contact_persons", contactPersonsController)
	controllers.RegisterController("documents", documentsController)
	controllers.RegisterController("dashboard", dashboardController)
	controllers.RegisterController("media", mediaController)
	controllers.RegisterController("public-events", publicEventsController)
	controllers.RegisterController("public", publicController)
	controllers.RegisterController("finance_accounts", financeAccountsController)
	controllers.RegisterController("bill_particulars", billParticularsController)
	controllers.RegisterController("bills", billsController)
	controllers.RegisterController("chat", chatController)
	controllers.RegisterController("payments", paymentsController)
	controllers.RegisterController("finance-reports", financeReportsController)
	controllers.RegisterController("sms", smsController)
	controllers.RegisterController("activity_logs", activityLogsController)
	controllers.RegisterController("school-bills", schoolBillsController)
	controllers.RegisterController("school-payments", schoolPaymentsController)
}
