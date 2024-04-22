package echopprof

import (
	"net/http/pprof"

	"github.com/rs/zerolog/log"

	"github.com/labstack/echo/v4"
)

func Wrap(e *echo.Echo) error {
	wrapGroup(e.Group("/debug"))
	return nil
}

func wrapGroup(g *echo.Group) error {
	routers := []struct {
		method  string
		path    string
		handler echo.HandlerFunc
	}{
		{"GET", "/pprof", IndexHandler()},
		{"GET", "/heap", HeapHandler()},
		{"GET", "/block", BlockHandler()},
		{"GET", "/allocs", AllocHandler()},
		{"GET", "/cmdline", CmdlineHandler()},
		{"GET", "/goroutine", GoroutineHandler()},
		{"GET", "/mutex", MutexHandler()},
		{"GET", "/profile", ProfileHandler()},
		{"GET", "/threadcreate", ThreadHandler()},
		{"GET", "/trace", TraceHandler()},
	}
	for _, r := range routers {
		switch r.method {
		case "GET":
			g.GET(r.path, r.handler)
		}
	}
	return nil
}

func IndexHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Info().Msg("Profiling index")
		pprof.Index(c.Response().Writer, c.Request())
		return nil
	}
}

func HeapHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Info().Msg("Profiling heap")
		pprof.Handler("heap").ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}
}

func BlockHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Info().Msg("Profiling block")
		pprof.Handler("block").ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}
}

func AllocHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Info().Msg("Profiling allocs")
		pprof.Handler("allocs").ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}
}

func CmdlineHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Info().Msg("Profiling cmdline")
		pprof.Cmdline(c.Response().Writer, c.Request())
		return nil
	}
}

func GoroutineHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Info().Msg("Profiling goroutine")
		pprof.Handler("goroutine").ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}
}

func MutexHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Info().Msg("Profiling mutex")
		pprof.Handler("mutex").ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}
}

func ProfileHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Info().Msg("Profiling profile")
		pprof.Profile(c.Response().Writer, c.Request())
		return nil
	}
}

func ThreadHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Info().Msg("Profiling threadcreate")
		pprof.Handler("threadcreate").ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}
}

func TraceHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Info().Msg("Profiling trace")
		pprof.Trace(c.Response().Writer, c.Request())
		return nil
	}
}
