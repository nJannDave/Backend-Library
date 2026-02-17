package wiring

import (
	pgc "stmnplibrary/controller/postgres/config"
	rdc "stmnplibrary/controller/redis/config"
	ra "stmnplibrary/controller/repository/admin"
	ru "stmnplibrary/controller/repository/user"
	rau "stmnplibrary/controller/repository/auth"
	sa "stmnplibrary/controller/service/admin"
	su "stmnplibrary/controller/service/user"
	sau "stmnplibrary/controller/service/auth"
	ha "stmnplibrary/controller/handler/admin"
	hu "stmnplibrary/controller/handler/user"
	hau "stmnplibrary/controller/handler/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func initializeApp() (*gin.Engine, func(), error) {
	wire.Build(
		pgc.ProviderConnStr,
		pgc.Init,
		rdc.ProviderCTX,
		rdc.ConnectRedis,
		ra.FnAdminRepository,
		ru.FnUserRepository,
		rau.FnAuthRepository,
		sa.FnAdminService,
		su.FnUserService,
		sau.FnAuthService,
		ha.FnAdminHandler,
		hu.FnUserHandler,
		hau.FnAuthHandler,
		WireHandler,
	)
	return nil, nil, nil
}