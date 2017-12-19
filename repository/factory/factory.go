package factory

import (
	"github.com/fireworq/fireworq/config"
	"github.com/fireworq/fireworq/repository"
	"github.com/fireworq/fireworq/repository/inmemory"
	"github.com/fireworq/fireworq/repository/mysql"

	"github.com/rs/zerolog/log"
)

// NewRepositories creates a new repository.Repositories instance
// according to the value of "driver" configuration.
func NewRepositories() *repository.Repositories {
	driver := config.Get("driver")

	var impl *repository.Repositories
	if driver == "mysql" {
		log.Info().Msg("Select mysql as a driver for repositories")
		db, err := mysql.NewDB()
		if err != nil {
			log.Panic().Msg(err.Error())
		}

		impl = &repository.Repositories{
			Queue:   mysql.NewQueueRepository(db),
			Routing: mysql.NewRoutingRepository(db),
		}
	}
	if driver == "in-memory" {
		log.Info().Msg("Select in-memory as a driver for repositories")
		impl = &repository.Repositories{
			Queue:   inmemory.NewQueueRepository(),
			Routing: inmemory.NewRoutingRepository(),
		}
	}

	if impl == nil {
		log.Panic().Msgf("Unknown driver: %s", driver)
	}

	return impl
}
