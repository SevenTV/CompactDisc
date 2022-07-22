package instance

import (
	"github.com/seventv/common/mongo"
	"github.com/seventv/common/redis"
	"github.com/seventv/common/structures/v3/query"
	"github.com/seventv/compactdisc/internal/discord"
)

type Instances struct {
	Mongo   mongo.Instance
	Redis   redis.Instance
	Discord discord.Instance

	Query *query.Query
}
