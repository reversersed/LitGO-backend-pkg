package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/gin-gonic/gin"
	shared_pb "github.com/reversersed/LitGO-proto/gen/go/shared"
	users_pb "github.com/reversersed/LitGO-proto/gen/go/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

//go:generate mockgen -source=JwtHandler.go -destination=mocks/jwt_mw_mock.go

const (
	TokenCookieName   string = "authTokenCookie"
	RefreshCookieName string = "refreshTokenCookie"
	UserIdKey         string = "userAuthId"
	UserLoginKey      string = "userLoginCredential"
	UserRolesKey      string = "userRolesCredential"
)

type Logger interface {
	Infof(string, ...any)
	Info(...any)
	Errorf(string, ...any)
	Error(...any)
	Warnf(string, ...any)
	Warn(...any)
}
type UserServer interface {
	UpdateToken(context.Context, *users_pb.TokenRequest, ...grpc.CallOption) (*users_pb.TokenReply, error)
}
type jwtMiddleware struct {
	secret     string
	logger     Logger
	userServer users_pb.UserClient
}
type claims struct {
	jwt.RegisteredClaims
	Login string   `json:"login"`
	Roles []string `json:"roles"`
	Email string   `json:"email"`
}
type UserTokenModel struct {
	Id    string   `json:"-"`
	Login string   `json:"login"`
	Roles []string `json:"roles"`
	Email string   `json:"-"`
}

func NewJwtMiddleware(logger Logger, secret string, userService users_pb.UserClient) (*jwtMiddleware, error) {

	return &jwtMiddleware{
		secret:     secret,
		logger:     logger,
		userServer: userService,
	}, nil
}

func (j *jwtMiddleware) Middleware(c *gin.Context) {
	headertoken, err := c.Cookie(TokenCookieName)
	if err != nil {
		return
	}
	key := []byte(j.secret)
	verifier, err := jwt.NewVerifierHS(jwt.HS256, key)
	if err != nil {
		j.logger.Errorf("error creating verifier for key. key length = %d, error = %v", len(key), err)
		c.Error(status.Error(codes.Unauthenticated, "error creating verifier for key"))
		c.Abort()
		return
	}
	j.logger.Info("parsing and verifying token...")
	token, err := jwt.ParseAndVerifyString(headertoken, verifier)
	if err != nil {
		j.logger.Errorf("error verifying token. error = %v", err)
		c.Error(status.Error(codes.Unauthenticated, "error verifying token"))
		c.Abort()
		return
	}

	var claims claims
	if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
		j.logger.Errorf("error unmarashalling claims: %v", err)
		c.Error(status.Error(codes.Unauthenticated, "error getting claims"))
		c.Abort()
		return
	}

	if !claims.IsValidAt(time.Now()) {
		refreshCookie, err := c.Cookie(RefreshCookieName)
		if err != nil {
			c.SetCookie(TokenCookieName, "", -1, "/", "", true, true)
			c.SetCookie(RefreshCookieName, "", -1, "/", "", true, true)
			c.Error(status.Error(codes.Unauthenticated, err.Error()))
			c.Abort()
			return
		}
		tokenReply, err := j.userServer.UpdateToken(c.Request.Context(), &users_pb.TokenRequest{Refreshtoken: refreshCookie})
		if err != nil {
			c.SetCookie(TokenCookieName, "", -1, "/", "", true, true)
			c.SetCookie(RefreshCookieName, "", -1, "/", "", true, true)
			c.Error(err)
			c.Abort()
			return
		}
		c.SetCookie(TokenCookieName, tokenReply.GetToken(), (int)((20*time.Minute)/time.Second), "/", "", true, true)
		c.SetCookie(RefreshCookieName, tokenReply.GetRefreshtoken(), (int)((31*24*time.Hour)/time.Second), "/", "", true, true)
	}

	j.logger.Infof("user's %s token has been verified with %v rights", claims.Login, claims.Roles)
	md := metadata.New(nil)
	md.Append(UserIdKey, claims.ID)
	md.Append(UserLoginKey, claims.Login)
	for _, role := range claims.Roles {
		md.Append(UserRolesKey, role)
	}
	ctx := metadata.NewIncomingContext(c.Request.Context(), md)
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}
func GetCredentialsFromContext(c context.Context, logger Logger) (*shared_pb.UserCredentials, error) {
	md, ok := metadata.FromIncomingContext(c)
	if !ok {
		return nil, status.New(codes.Unauthenticated, "no metadata credentials found").Err()
	}
	userId := md.Get(UserIdKey)
	if len(userId) != 1 {
		erro, _ := status.New(codes.Unauthenticated, "no user credentials found").WithDetails(&shared_pb.ErrorDetail{Field: "User ID", Description: "User id was not found in metadata"})
		return nil, erro.Err()
	}
	userLogin := md.Get(UserLoginKey)
	if len(userLogin) != 1 {
		logger.Warnf("can't get user %s login", userId[0])
		erro, _ := status.New(codes.Unauthenticated, "no user credentials found").WithDetails(&shared_pb.ErrorDetail{Field: "User Login", Description: "User login was not found in metadata"})
		return nil, erro.Err()
	}
	userRoles := md.Get(UserRolesKey)
	if len(userRoles) == 0 {
		logger.Warnf("can't get user %s %s roles", userId[0], userLogin[0])
		erro, _ := status.New(codes.Unauthenticated, "no user credentials found").WithDetails(&shared_pb.ErrorDetail{Field: "User Roles", Description: "User roles was not found in metadata"})
		return nil, erro.Err()
	}

	if logger != nil {
		logger.Infof("got user credentials: %s, %s, %v", userId[0], userLogin[0], userRoles)
	}
	return &shared_pb.UserCredentials{
		Id:    userId[0],
		Login: userLogin[0],
		Roles: userRoles,
	}, nil
}
func CreateTokenCookie(token string, refreshToken string, rememberMe bool) (tokenCookie http.Cookie, refreshCookie http.Cookie) {
	if len(token) == 0 {
		tokenCookie = http.Cookie{
			Name:     TokenCookieName,
			Value:    "",
			MaxAge:   -1,
			Path:     "/",
			Domain:   "",
			Secure:   true,
			HttpOnly: true,
		}
	} else {
		tokenCookie = http.Cookie{
			Name:     TokenCookieName,
			Value:    token,
			MaxAge:   (int)((20 * time.Minute) / time.Second),
			Path:     "/",
			Domain:   "",
			Secure:   true,
			HttpOnly: true,
		}
	}
	if len(refreshToken) == 0 {
		refreshCookie = http.Cookie{
			Name:     RefreshCookieName,
			Value:    "",
			MaxAge:   -1,
			Path:     "/",
			Domain:   "",
			Secure:   true,
			HttpOnly: true,
		}
	} else {
		refreshCookie = http.Cookie{
			Name:     RefreshCookieName,
			Value:    refreshToken,
			MaxAge:   (int)((31 * 24 * time.Hour) / time.Second),
			Path:     "/",
			Domain:   "",
			Secure:   true,
			HttpOnly: true,
		}
	}
	return
}
