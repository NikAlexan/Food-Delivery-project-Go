package proxy

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "food-delivery/delivery-service/proto/pb"
)

type DeliveryProxy struct {
	conn   *grpc.ClientConn
	client pb.DeliveryServiceClient
}

func NewDeliveryProxy(deliveryServiceAddr string) (*DeliveryProxy, error) {
	conn, err := grpc.NewClient(deliveryServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &DeliveryProxy{conn: conn, client: pb.NewDeliveryServiceClient(conn)}, nil
}

func (p *DeliveryProxy) Close() error {
	return p.conn.Close()
}

func (p *DeliveryProxy) AssignDriver(c *gin.Context) {
	var req pb.AssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := p.client.AssignDriver(userContext(c), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (p *DeliveryProxy) GetDelivery(c *gin.Context) {
	deliveryID, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}

	resp, err := p.client.GetDelivery(userContext(c), &pb.DeliveryIdRequest{DeliveryId: deliveryID})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *DeliveryProxy) UpdateDriverLocation(c *gin.Context) {
	var req pb.LocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := p.client.UpdateDriverLocation(userContext(c), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (p *DeliveryProxy) TrackDelivery(c *gin.Context) {
	deliveryID, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}

	resp, err := p.client.TrackDelivery(userContext(c), &pb.DeliveryIdRequest{DeliveryId: deliveryID})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *DeliveryProxy) CompleteDelivery(c *gin.Context) {
	deliveryID, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}

	resp, err := p.client.CompleteDelivery(userContext(c), &pb.DeliveryIdRequest{DeliveryId: deliveryID})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *DeliveryProxy) ListDriverDeliveries(c *gin.Context) {
	driverID, ok := parseInt64Param(c, "driverId")
	if !ok {
		return
	}

	resp, err := p.client.ListDriverDeliveries(userContext(c), &pb.DriverIdRequest{DriverId: driverID})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *DeliveryProxy) GetDeliveryHistory(c *gin.Context) {
	userID := c.GetInt64("user_id")
	resp, err := p.client.GetDeliveryHistory(userContext(c), &pb.UserIdRequest{UserId: userID})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *DeliveryProxy) CancelDelivery(c *gin.Context) {
	deliveryID, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}

	resp, err := p.client.CancelDelivery(userContext(c), &pb.DeliveryIdRequest{DeliveryId: deliveryID})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *DeliveryProxy) RegisterDriver(c *gin.Context) {
	var req pb.RegisterDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := p.client.RegisterDriver(userContext(c), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (p *DeliveryProxy) GetMyDriver(c *gin.Context) {
	resp, err := p.client.GetMyDriver(userContext(c), &pb.Empty{})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *DeliveryProxy) SetAvailability(c *gin.Context) {
	var req pb.AvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	resp, err := p.client.SetAvailability(userContext(c), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func parseInt64Param(c *gin.Context, name string) (int64, bool) {
	value, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path parameter"})
		return 0, false
	}
	return value, true
}

func userContext(c *gin.Context) context.Context {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		return c.Request.Context()
	}

	return metadata.AppendToOutgoingContext(
		c.Request.Context(),
		"x-user-id", strconv.FormatInt(userID, 10),
	)
}
