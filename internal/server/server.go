package server

import (
	"context"
	goerrors "errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	_ "lumina/docs"
	"lumina/internal/config"
	"lumina/pkg/log"
)

const httpXRequestId = "X-Request-Id"

type Server struct {
	conf       *config.Config
	httpServer *http.Server
	client     *http.Client
	logger     *logrus.Entry
}

func NewServer(ctx context.Context, conf *config.Config) (*Server, error) {
	s := &Server{
		conf: conf,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		logger: log.GetLogger(ctx),
	}

	return s, nil
}

func RequestId() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := c.GetHeader(httpXRequestId)
		if requestId == "" {
			requestId = strings.ReplaceAll(uuid.New().String(), "-", "")
		}
		c.Header(httpXRequestId, requestId)
		c.Next()
	}
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		latency := time.Since(t)
		status := c.Writer.Status()

		var totalTokens int
		if val, exists := c.Get("total_tokens"); exists {
			totalTokens = val.(int)
		}

		logrus.Info("ip: ", c.ClientIP(), " method: ", c.Request.Method, " path: ",
			c.Request.URL.Path, " status: ", status, " latency: ", latency, " total_tokens: ", totalTokens)
	}
}

func (s *Server) Start() {
	gin.SetMode(gin.ReleaseMode)
	router := s.SetUpRouter()
	pprof.Register(router)
	s.httpServer = &http.Server{
		Addr:    s.conf.Addr,
		Handler: router,
	}

	var err error
	if s.conf.SSLCert != "" && s.conf.SSLKey != "" {
		logrus.Infof("start https server on %s", s.conf.Addr)
		err = s.httpServer.ListenAndServeTLS(s.conf.SSLCert, s.conf.SSLKey)
	} else {
		logrus.Infof("start http server on %s", s.conf.Addr)
		err = s.httpServer.ListenAndServe()
	}
	if err != nil && !goerrors.Is(err, http.ErrServerClosed) {
		logrus.Fatal(err)
	}
}

func (s *Server) Shutdown() {
	err := s.httpServer.Shutdown(context.Background())
	if err != nil {
		logrus.Fatalf("server forced to shutdown: %v", err)
	}
}

type ErrorResponse struct {
	// 错误信息
	Error string `json:"error"`
}

func (s *Server) writeError(c *gin.Context, code int, err error) {
	c.JSON(code, ErrorResponse{
		Error: err.Error(),
	})
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("password", func(fl validator.FieldLevel) bool {
			matched, _ := regexp.MatchString(`^[a-zA-Z0-9!@#$%^*+()]+$`, fl.Field().String())
			return matched
		})
	}
}
