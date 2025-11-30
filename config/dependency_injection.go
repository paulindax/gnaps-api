package config

import (
	"gnaps-api/controllers"
	"gnaps-api/repositories"
	"gnaps-api/services"

	"gorm.io/gorm"
)

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

	// Initialize Services
	eventService := services.NewEventService(eventRepo, registrationRepo)
	schoolService := services.NewSchoolService(schoolRepo)
	newsService := services.NewNewsService(newsRepo, commentRepo, userRepo)
	regionService := services.NewRegionService(regionRepo)
	zoneService := services.NewZoneService(zoneRepo)
	groupService := services.NewGroupService(groupRepo)
	positionService := services.NewPositionService(positionRepo)
	executiveService := services.NewExecutiveService(executiveRepo)
	contactPersonService := services.NewContactPersonService(contactPersonRepo)
	documentService := services.NewDocumentService(documentRepo)
	dashboardService := services.NewDashboardService(db)
	mediaService := services.NewMediaService()

	// Initialize Controllers
	publicEventsController := controllers.NewPublicEventsController(eventRepo, registrationRepo, schoolRepo)

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
}
