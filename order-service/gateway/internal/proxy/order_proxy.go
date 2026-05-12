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

func NewOrderProxy(orderServiceAddr string) (*OrderProxy, error) {
	conn, err := grpc.NewClient(orderServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &OrderProxy{conn: conn, client: pb.NewOrderServiceClient(conn)}, nil
}

func (p *OrderProxy) Close() error {
	return p.conn.Close()
}

func respondGRPCError(c *gin.Context, err error) {
	code, msg := grpcError(err)
	c.JSON(code, gin.H{"error": msg})
}

// POST /api/orders
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

// GET /api/orders/:id
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

// GET /api/orders
func (p *OrderProxy) ListUserOrders(c *gin.Context) {
	userID := c.GetInt64("user_id")
	resp, err := p.client.ListUserOrders(c.Request.Context(), &pb.UserIdRequest{UserId: userID})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// PATCH /api/orders/:id/status
func (p *OrderProxy) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	var req pb.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.OrderId = id

	resp, err := p.client.UpdateOrderStatus(c.Request.Context(), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// POST /api/orders/:id/cancel
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

// POST /api/orders/:id/pay
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

// GET /api/orders/history
func (p *OrderProxy) GetOrderHistory(c *gin.Context) {
	userID := c.GetInt64("user_id")
	resp, err := p.client.GetOrderHistory(c.Request.Context(), &pb.UserIdRequest{UserId: userID})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// POST /api/orders/total
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
