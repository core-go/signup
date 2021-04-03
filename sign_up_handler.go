package signup

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

type SignUpHandler struct {
	SignUpService SignUpService
	ErrorStatus   int
	Error         func(context.Context, string)
	Decrypt       func(cipherText string, secretKey string) (string, error)
	EncryptionKey string
	Log           func(ctx context.Context, resource string, action string, success bool, desc string) error
	Config        SignUpActionConfig
}

func NewSignUpHandlerWithDecrypter(signUpService SignUpService, errorStatus int, logError func(context.Context, string), decrypt func(cipherText string, secretKey string) (string, error), encryptionKey string, conf *SignUpActionConfig, options...func(context.Context, string, string, bool, string) error) *SignUpHandler {
	var c SignUpActionConfig
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
	return &SignUpHandler{SignUpService: signUpService, ErrorStatus: errorStatus, Config: c, Error: logError, Log: writeLog, Decrypt: decrypt, EncryptionKey: encryptionKey}
}

func NewSignUpHandler(signUpService SignUpService, errorStatus int, logError func(context.Context, string), conf *SignUpActionConfig, options...func(context.Context, string, string, bool, string) error) *SignUpHandler {
	return NewSignUpHandlerWithDecrypter(signUpService, errorStatus, logError, nil, "", conf, options...)
}

func (h *SignUpHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	ip := getRemoteIp(r)
	var ctx context.Context
	ctx = r.Context()
	if len(h.Config.Ip) > 0 {
		ctx = context.WithValue(ctx, h.Config.Ip, ip)
		r = r.WithContext(ctx)
	}
	var user SignUpInfo
	er1 := json.NewDecoder(r.Body).Decode(&user)
	if er1 != nil {
		if h.Error != nil {
			h.Error(r.Context(), "Cannot decode sign up info: "+er1.Error())
		}
		http.Error(w, "Cannot decode sign up info", http.StatusBadRequest)
		return
	}
	if h.Decrypt != nil && len(h.EncryptionKey) > 0 {
		decodedPassword, er2 := h.Decrypt(user.Password, h.EncryptionKey)
		if er2 != nil {
			if h.Error != nil {
				msg := "cannot decode password: " + er2.Error()
				h.Error(r.Context(), msg)
			}
			http.Error(w, "cannot decode password", http.StatusBadRequest)
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
		respond(w, r, http.StatusOK, result, h.Log, h.Config.Resource, h.Config.Signup, false, msg)
	} else {
		respond(w, r, http.StatusOK, result, h.Log, h.Config.Resource, h.Config.Signup, true, "")
	}
}
func (h *SignUpHandler) VerifyUser(w http.ResponseWriter, r *http.Request) {
	s := strings.Split(r.RequestURI, "/")
	if len(s) < 3 {
		http.Error(w, "URI is invalid", http.StatusBadRequest)
		return
	}
	userId := s[len(s)-2]
	code := s[len(s)-1]
	if len(userId) == 0 {
		http.Error(w, "User Id is required to verify user", http.StatusBadRequest)
		return
	}
	if len(code) == 0 {
		http.Error(w, "Verified code is required to verify user", http.StatusBadRequest)
		return
	}
	result, err := h.SignUpService.VerifyUser(r.Context(), userId, code)
	if err != nil {
		msg := err.Error()
		if h.Error != nil {
			h.Error(r.Context(), msg)
		}
		respond(w, r, http.StatusInternalServerError, "cannot verify user", h.Log, h.Config.Resource, h.Config.VerifyUser, false, msg)
	} else {
		respond(w, r, http.StatusOK, result, h.Log, h.Config.Resource, h.Config.VerifyUser, true, "")
	}
}
func (h *SignUpHandler) VerifyUserAndSavePassword(w http.ResponseWriter, r *http.Request) {
	var user VerificationInfo
	er1 := json.NewDecoder(r.Body).Decode(&user)
	if er1 != nil {
		if h.Error != nil {
			h.Error(r.Context(), "Cannot decode verification info: "+er1.Error())
		}
		http.Error(w, "Cannot decode verification info", http.StatusBadRequest)
		return
	}
	if len(user.Id) == 0 {
		http.Error(w, "User Id is required to verify user", http.StatusBadRequest)
		return
	}
	if len(user.Passcode) == 0 {
		http.Error(w, "Verified code is required to verify user", http.StatusBadRequest)
		return
	}
	if len(user.Password) == 0 {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}
	result, err := h.SignUpService.VerifyUserAndSavePassword(r.Context(), user.Id, user.Passcode, user.Password)
	if err != nil {
		msg := err.Error()
		if h.Error != nil {
			h.Error(r.Context(), msg)
		}
		respond(w, r, http.StatusInternalServerError, "cannot verify user and save password", h.Log, h.Config.Resource, h.Config.VerifyUser, false, msg)
	} else {
		respond(w, r, http.StatusOK, result, h.Log, h.Config.Resource, h.Config.VerifyUser, true, "")
	}
}
func respond(w http.ResponseWriter, r *http.Request, code int, result interface{}, writeLog func(context.Context, string, string, bool, string) error, resource string, action string, success bool, desc string) {
	response, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
	if writeLog != nil {
		writeLog(r.Context(), resource, action, success, desc)
	}
}
func getRemoteIp(r *http.Request) string {
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteIP = r.RemoteAddr
	}
	return remoteIP
}
