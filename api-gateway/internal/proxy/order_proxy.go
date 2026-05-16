package proxy

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "food-delivery/order-service/proto/pb"
)

type OrderProxy struct {
	conn   *grpc.ClientConn
	client pb.OrderServiceClient
}

func NewOrderProxy(addr string) (*OrderProxy, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &OrderProxy{conn: conn, client: pb.NewOrderServiceClient(conn)}, nil
}

func (p *OrderProxy) Close() error {
	return p.conn.Close()
}

func (p *OrderProxy) CreateOrder(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req pb.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.UserId = userID

	resp, err := p.client.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (p *OrderProxy) GetOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	resp, err := p.client.GetOrder(c.Request.Context(), &pb.OrderIdRequest{OrderId: id})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *OrderProxy) ListUserOrders(c *gin.Context) {
	userID := c.GetInt64("user_id")
	resp, err := p.client.ListUserOrders(c.Request.Context(), &pb.UserIdRequest{UserId: userID})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *OrderProxy) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := p.client.UpdateOrderStatus(c.Request.Context(), &pb.UpdateStatusRequest{
		OrderId: id,
		Status:  body.Status,
	})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *OrderProxy) CancelOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	resp, err := p.client.CancelOrder(c.Request.Context(), &pb.OrderIdRequest{OrderId: id})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *OrderProxy) ProcessPayment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	var req pb.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.OrderId = id

	resp, err := p.client.ProcessPayment(c.Request.Context(), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *OrderProxy) GetOrderHistory(c *gin.Context) {
	userID := c.GetInt64("user_id")
	resp, err := p.client.GetOrderHistory(c.Request.Context(), &pb.UserIdRequest{UserId: userID})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (p *OrderProxy) CalculateTotal(c *gin.Context) {
	var req pb.CartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := p.client.CalculateTotal(c.Request.Context(), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
