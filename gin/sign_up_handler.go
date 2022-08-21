package gin

import (
	"context"
	"encoding/json"
	s "github.com/core-go/signup"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"strings"
)

type SignUpHandler struct {
	SignUpService s.SignUpService
	ErrorStatus   int
	Error         func(context.Context, string, ...map[string]interface{})
	Decrypt       func(string) (string, error)
	Log           func(ctx context.Context, resource string, action string, success bool, desc string) error
	Config        s.SignUpActionConfig
}

func NewSignUpHandlerWithDecrypter(signUpService s.SignUpService, errorStatus int, logError func(context.Context, string, ...map[string]interface{}), decrypt func(cipherText string) (string, error), conf *s.SignUpActionConfig, options...func(context.Context, string, string, bool, string) error) *SignUpHandler {
	var c s.SignUpActionConfig
	if conf != nil {
		c.Resource = conf.Resource
		c.Signup = conf.Signup
		c.VerifyUser = conf.VerifyUser
	}
	if len(c.Resource) == 0 {
		c.Resource = "signup"
	}
	if len(c.Signup) == 0 {
		c.Signup = "signup"
	}
	if len(c.VerifyUser) == 0 {
		c.VerifyUser = "verify"
	}
	if len(c.Ip) == 0 {
		c.Ip = "ip"
	}
	var writeLog func(context.Context, string, string, bool, string) error
	if len(options) >= 1 {
		writeLog = options[0]
	}
	return &SignUpHandler{SignUpService: signUpService, ErrorStatus: errorStatus, Config: c, Error: logError, Log: writeLog, Decrypt: decrypt}
}

func NewSignUpHandler(signUpService s.SignUpService, errorStatus int, logError func(context.Context, string, ...map[string]interface{}), conf *s.SignUpActionConfig, options...func(context.Context, string, string, bool, string) error) *SignUpHandler {
	return NewSignUpHandlerWithDecrypter(signUpService, errorStatus, logError, nil, conf, options...)
}

func (h *SignUpHandler) SignUp(ctx *gin.Context) {
	ip := getRemoteIp(ctx.Request)
	var ctx2 context.Context
	ctx2 = ctx.Request.Context()
	if len(h.Config.Ip) > 0 {
		ctx2 = context.WithValue(ctx2, h.Config.Ip, ip)
		ctx.Request = ctx.Request.WithContext(ctx2)
	}
	r := ctx.Request
	var user s.SignUpInfo
	er1 := json.NewDecoder(r.Body).Decode(&user)
	if er1 != nil {
		if h.Error != nil {
			h.Error(r.Context(), "Cannot decode sign up info: "+er1.Error())
		}
		ctx.String(http.StatusBadRequest, "Cannot decode sign up info")
		return
	}
	if h.Decrypt != nil {
		decodedPassword, er2 := h.Decrypt(user.Password)
		if er2 != nil {
			if h.Error != nil {
				msg := "cannot decode password: " + er2.Error()
				h.Error(r.Context(), msg)
			}
			ctx.String(http.StatusBadRequest, "cannot decode password")
			return
		}
		user.Password = decodedPassword
	}
	result, er3 := h.SignUpService.SignUp(r.Context(), user)
	if er3 != nil {
		result.Status = h.ErrorStatus
		msg := er3.Error()
		if h.Error != nil {
			h.Error(r.Context(), msg)
		}
		respond(ctx, http.StatusOK, result, h.Log, h.Config.Resource, h.Config.Signup, false, msg)
	} else {
		respond(ctx, http.StatusOK, result, h.Log, h.Config.Resource, h.Config.Signup, true, "")
	}
}
func (h *SignUpHandler) VerifyUser(ctx *gin.Context) {
	r := ctx.Request
	s := strings.Split(r.RequestURI, "/")
	if len(s) < 3 {
		ctx.String(http.StatusBadRequest, "URI is invalid")
		return
	}
	userId := s[len(s)-2]
	code := s[len(s)-1]
	if len(userId) == 0 {
		ctx.String(http.StatusBadRequest, "User Id is required to verify user")
		return
	}
	if len(code) == 0 {
		ctx.String(http.StatusBadRequest, "Verified code is required to verify user")
		return
	}
	result, err := h.SignUpService.VerifyUser(r.Context(), userId, code)
	if err != nil {
		msg := err.Error()
		if h.Error != nil {
			h.Error(r.Context(), msg)
		}
		respond(ctx, http.StatusInternalServerError, "cannot verify user", h.Log, h.Config.Resource, h.Config.VerifyUser, false, msg)
	} else {
		respond(ctx, http.StatusOK, result, h.Log, h.Config.Resource, h.Config.VerifyUser, true, "")
	}
}
func (h *SignUpHandler) VerifyUserAndSavePassword(ctx *gin.Context) {
	r := ctx.Request
	var user s.VerificationInfo
	er1 := json.NewDecoder(r.Body).Decode(&user)
	if er1 != nil {
		if h.Error != nil {
			h.Error(r.Context(), "Cannot decode verification info: "+er1.Error())
		}
		ctx.String(http.StatusBadRequest, "Cannot decode verification info")
		return
	}
	if len(user.Id) == 0 {
		ctx.String(http.StatusBadRequest, "User Id is required to verify user")
		return
	}
	if len(user.Passcode) == 0 {
		ctx.String(http.StatusBadRequest, "Verified code is required to verify user")
		return
	}
	if len(user.Password) == 0 {
		ctx.String(http.StatusBadRequest, "Password is required")
		return
	}
	result, err := h.SignUpService.VerifyUserAndSavePassword(r.Context(), user.Id, user.Passcode, user.Password)
	if err != nil {
		msg := err.Error()
		if h.Error != nil {
			h.Error(r.Context(), msg)
		}
		respond(ctx, http.StatusInternalServerError, "cannot verify user and save password", h.Log, h.Config.Resource, h.Config.VerifyUser, false, msg)
	} else {
		respond(ctx, http.StatusOK, result, h.Log, h.Config.Resource, h.Config.VerifyUser, true, "")
	}
}
func respond(ctx *gin.Context, code int, result interface{}, writeLog func(context.Context, string, string, bool, string) error, resource string, action string, success bool, desc string) {
	ctx.JSON(code, result)
	if writeLog != nil {
		writeLog(ctx.Request.Context(), resource, action, success, desc)
	}
}
func getRemoteIp(r *http.Request) string {
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteIP = r.RemoteAddr
	}
	return remoteIP
}
