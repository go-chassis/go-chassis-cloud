package main

import (
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/v2"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"net/http"

	_ "github.com/go-chassis/go-chassis-cloud/provider/huawei/engine"
)

func main() {
	chassis.RegisterSchema("rest", &HelloResource{})
	err := chassis.Init()
	if err != nil {
		panic(err)
	}
	err = chassis.Run()
	if err != nil {
		panic(err)
	}
}

type HelloResource struct {
}

func (r *HelloResource) Welcome(b *restful.Context) {
	b.Write([]byte("hello: " + archaius.GetString("user", "")))

}

func (r *HelloResource) URLPatterns() []restful.Route {
	return []restful.Route{
		{Method: http.MethodPost, Path: "/welcome", ResourceFunc: r.Welcome,
			Returns: []*restful.Returns{{Code: 200}}},
	}
}
