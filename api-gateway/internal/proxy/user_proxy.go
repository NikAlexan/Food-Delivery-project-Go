package proxy

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "food-delivery/user-service/proto/pb"
)

type UserProxy struct {
	conn   *grpc.ClientConn
	client pb.UserServiceClient
}

func NewUserProxy(userServiceAddr string) (*UserProxy, error) {
	conn, err := grpc.NewClient(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &UserProxy{conn: conn, client: pb.NewUserServiceClient(conn)}, nil
}

func (p *UserProxy) Close() error {
	return p.conn.Close()
}

func respondGRPCError(c *gin.Context, err error) {
	code, msg := grpcError(err)
	c.JSON(code, gin.H{"error": msg})
}

func (p *UserProxy) Register(c *gin.Context) {
	var req pb.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := p.client.RegisterUser(c.Request.Context(), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (p *UserProxy) Login(c *gin.Context) {
	var req pb.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := p.client.LoginUser(c.Request.Context(), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *UserProxy) GetProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	resp, err := p.client.GetProfile(c.Request.Context(), &pb.UserIdRequest{UserId: userID})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *UserProxy) UpdateProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req pb.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.UserId = userID

	resp, err := p.client.UpdateProfile(c.Request.Context(), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *UserProxy) AddAddress(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req pb.AddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.UserId = userID

	resp, err := p.client.AddAddress(c.Request.Context(), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (p *UserProxy) GetAddresses(c *gin.Context) {
	userID := c.GetInt64("user_id")
	resp, err := p.client.GetAddresses(c.Request.Context(), &pb.UserIdRequest{UserId: userID})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *UserProxy) RefreshToken(c *gin.Context) {
	var req pb.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := p.client.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *UserProxy) DeleteUser(c *gin.Context) {
	userID := c.GetInt64("user_id")
	_, err := p.client.DeleteUser(c.Request.Context(), &pb.UserIdRequest{UserId: userID})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
