package http

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/soumajitgh/mobicode/internal/graphql/generated"
	"github.com/soumajitgh/mobicode/internal/graphql/resolver"
)

func newGraphQLHandler(resolver *resolver.Resolver) *handler.Server {
	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.POST{})
	srv.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		presented := graphql.DefaultErrorPresenter(ctx, err)
		if presented.Extensions == nil {
			presented.Extensions = map[string]any{}
		}
		if _, ok := presented.Extensions["code"]; !ok {
			code := "VALIDATION_FAILED"
			if presented.Message == "internal system error" {
				code = "INTERNAL_ERROR"
			}
			presented.Extensions["code"] = code
		}
		return presented
	})
	return srv
}
