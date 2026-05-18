package proxy

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "food-delivery/restaurant-service/proto/pb"
)

// RestaurantProxy forwards HTTP requests to the restaurant-service gRPC backend.
type RestaurantProxy struct {
	conn   *grpc.ClientConn
	client pb.RestaurantServiceClient
}

// NewRestaurantProxy dials the restaurant-service and returns a ready proxy.
func NewRestaurantProxy(addr string) (*RestaurantProxy, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &RestaurantProxy{conn: conn, client: pb.NewRestaurantServiceClient(conn)}, nil
}

// Close releases the underlying gRPC connection.
func (p *RestaurantProxy) Close() error { return p.conn.Close() }

// ── Restaurants ───────────────────────────────────────────────────────────────

// CreateRestaurant handles POST /api/restaurants
func (p *RestaurantProxy) CreateRestaurant(c *gin.Context) {
	var req pb.CreateRestaurantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := p.client.CreateRestaurant(userContext(c), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// GetMyRestaurant handles GET /api/restaurants/my
func (p *RestaurantProxy) GetMyRestaurant(c *gin.Context) {
	resp, err := p.client.GetMyRestaurant(userContext(c), &pb.Empty{})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetRestaurant handles GET /api/restaurants/:id
func (p *RestaurantProxy) GetRestaurant(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}
	resp, err := p.client.GetRestaurant(c.Request.Context(), &pb.RestaurantIdRequest{RestaurantId: id})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListRestaurants handles GET /api/restaurants
func (p *RestaurantProxy) ListRestaurants(c *gin.Context) {
	categoryID, _ := strconv.ParseInt(c.Query("category_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := p.client.ListRestaurants(c.Request.Context(), &pb.ListRestaurantsRequest{
		CategoryId: categoryID,
		Page:       int32(page),
		PageSize:   int32(pageSize),
	})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SearchRestaurants handles GET /api/restaurants/search?q=...
func (p *RestaurantProxy) SearchRestaurants(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := p.client.SearchRestaurants(c.Request.Context(), &pb.SearchRequest{
		Query:    q,
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateRestaurant handles PUT /api/restaurants/:id
func (p *RestaurantProxy) UpdateRestaurant(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}
	var req pb.UpdateRestaurantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.RestaurantId = id

	resp, err := p.client.UpdateRestaurant(userContext(c), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteRestaurant handles DELETE /api/restaurants/:id
func (p *RestaurantProxy) DeleteRestaurant(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}
	_, err = p.client.DeleteRestaurant(userContext(c), &pb.RestaurantIdRequest{RestaurantId: id})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ── Menu ──────────────────────────────────────────────────────────────────────

// GetMenu handles GET /api/restaurants/:id/menu
func (p *RestaurantProxy) GetMenu(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}
	resp, err := p.client.GetMenu(c.Request.Context(), &pb.RestaurantIdRequest{RestaurantId: id})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateMenuItem handles POST /api/restaurants/:id/menu
func (p *RestaurantProxy) CreateMenuItem(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}
	var req pb.CreateMenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.RestaurantId = id

	resp, err := p.client.CreateMenuItem(userContext(c), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// UpdateMenuItem handles PUT /api/restaurants/:id/menu/:item_id
func (p *RestaurantProxy) UpdateMenuItem(c *gin.Context) {
	restaurantID, err := parseID(c, "id")
	if err != nil {
		return
	}
	itemID, err := parseID(c, "item_id")
	if err != nil {
		return
	}

	var req pb.UpdateMenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.RestaurantId = restaurantID
	req.ItemId = itemID

	resp, err := p.client.UpdateMenuItem(userContext(c), &req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteMenuItem handles DELETE /api/restaurants/:id/menu/:item_id
func (p *RestaurantProxy) DeleteMenuItem(c *gin.Context) {
	itemID, err := parseID(c, "item_id")
	if err != nil {
		return
	}
	_, err = p.client.DeleteMenuItem(userContext(c), &pb.MenuItemIdRequest{ItemId: itemID})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func parseID(c *gin.Context, param string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(param), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
		return 0, err
	}
	return id, nil
}
