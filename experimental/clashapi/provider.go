package clashapi

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing/common"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/json/badjson"
)

func proxyProviderRouter(server *Server) http.Handler {
	r := chi.NewRouter()
	r.Get("/", getProviders(server))

	r.Route("/{name}", func(r chi.Router) {
		r.Use(parseProviderName, findProviderByName(server))
		r.Get("/", getProvider(server))
		r.Put("/", updateProvider())
		r.Get("/healthcheck", nil)
	})
	return r
}

func parseProviderName(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := getEscapeParam(r, "name")
		ctx := context.WithValue(r.Context(), CtxKeyProviderName, name)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func findProviderByName(server *Server) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			name := r.Context().Value(CtxKeyProviderName).(string)
			provider, exist := server.provider.Provider(name)
			if !exist {
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, ErrNotFound)
				return
			}
			ctx := context.WithValue(r.Context(), CtxKeyProvider, provider)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func providerInfo(server *Server, provider adapter.Provider) *badjson.JSONObject {
	var outbounds []adapter.Outbound
	if manager := provider.Outbound(); manager != nil {
		outbounds = manager.Outbounds()
	}
	var info badjson.JSONObject
	info.Put("type", "Proxy")
	info.Put("name", provider.Tag())
	info.Put("vehicleType", provider.Type())
	info.Put("subscriptionInfo", provider.SubInfo())
	info.Put("updatedAt", provider.LastUpdateTime().Format("2006-01-02T15:04:05.999999999-07:00"))
	info.Put("proxies", common.Map(outbounds, func(it adapter.Outbound) *badjson.JSONObject {
		return proxyInfo(server, it)
	}))
	return &info
}

func getProviders(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var providerMap badjson.JSONObject
		providers := common.Filter(server.provider.Providers(), func(provider adapter.Provider) bool {
			return provider.Tag() != ""
		})

		for i, provider := range providers {
			var tag string
			if provider.Tag() == "" {
				tag = F.ToString(i)
			} else {
				tag = provider.Tag()
			}
			providerMap.Put(tag, providerInfo(server, provider))
		}
		var responseMap badjson.JSONObject
		if providerMap.IsEmpty() {
			responseMap.Put("providers", map[string]any{})
		} else {
			responseMap.Put("providers", &providerMap)
		}
		response, err := responseMap.MarshalJSON()
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, newError(err.Error()))
			return
		}
		w.Write(response)
	}
}

func getProvider(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := r.Context().Value(CtxKeyProvider).(adapter.Provider)
		response, err := providerInfo(server, provider).MarshalJSON()
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, newError(err.Error()))
			return
		}
		w.Write(response)
	}
}

func updateProvider() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := r.Context().Value(CtxKeyProvider).(adapter.Provider)
		err := provider.UpdateProvider()
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, err)
			return
		}
		render.NoContent(w, r)
	}
}

// func healthCheckProvider() func(w http.ResponseWriter, r *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		provider := r.Context().Value(CtxKeyProvider).(adapter.Provider)
// 		query := r.URL.Query()
// 		link := query.Get("url")
// 		timeout := int64(5000)
// 		ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*time.Duration(timeout))
// 		defer cancel()
// 		render.JSON(w, r, provider.Healthcheck(ctx, link))
// 	}
// }
