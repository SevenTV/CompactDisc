package instance

import (
	"github.com/seventv/common/mongo"
	"github.com/seventv/common/redis"
)

type Instances struct {
	Mongo mongo.Instance
	Redis redis.Instance
}
