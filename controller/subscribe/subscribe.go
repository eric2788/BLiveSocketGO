package subscribe

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/eric2788/biligo-live-ws/services/api"
	"github.com/eric2788/biligo-live-ws/services/subscriber"
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
	"time"
)

var Id = subscriber.ToClientId

func Register(gp *gin.RouterGroup) {
	gp.GET("", GetSubscriptions)
	gp.POST("", Subscribe)
	gp.DELETE("", ClearSubscribe)
	gp.PUT("add", AddSubscribe)
	gp.PUT("remove", RemoveSubscribe)
}

func GetSubscriptions(c *gin.Context) {
	list, ok := subscriber.Get(Id(c))
	if !ok {
		list = []int64{}
	}
	c.IndentedJSON(200, list)
}

func ClearSubscribe(c *gin.Context) {
	subscriber.Delete(Id(c))
	c.Status(200)
}

func AddSubscribe(c *gin.Context) {
	rooms, ok := GetSubscribesArr(c)

	if !ok {
		return
	}

	log.Printf("用戶 %v 新增訂閱 %v \n", Id(c), rooms)

	ActivateExpire(Id(c))

	newRooms := subscriber.Add(Id(c), rooms)
	c.IndentedJSON(200, newRooms)
}

func RemoveSubscribe(c *gin.Context) {
	rooms, ok := GetSubscribesArr(c)

	if !ok {
		return
	}

	log.Printf("用戶 %v 移除訂閱 %v \n", Id(c), rooms)

	newRooms, ok := subscriber.Remove(Id(c), rooms)

	if !ok {
		c.IndentedJSON(400, gin.H{"error": "刪除失敗，你尚未遞交過任何訂閱"})
		return
	}

	c.IndentedJSON(200, newRooms)
}

func Subscribe(c *gin.Context) {

	rooms, ok := GetSubscribesArr(c)

	if !ok {
		return
	}

	log.Printf("用戶 %v 設置訂閱 %v \n", Id(c), rooms)

	ActivateExpire(Id(c))

	subscriber.Update(Id(c), rooms)
	c.IndentedJSON(200, rooms)
}

func GetSubscribesArr(c *gin.Context) ([]int64, bool) {

	subArr, ok := c.GetPostFormArray("subscribes")
	if !ok {
		c.AbortWithStatusJSON(400, gin.H{"error": "缺少 `subscribes` 數值(訂閱列表)"})
		return nil, false
	}
	if len(subArr) == 0 {
		c.AbortWithStatusJSON(400, gin.H{"error": "訂閱列表不能為空"})
		return nil, false
	}

	roomSet := mapset.NewSet()

	for _, arr := range subArr {

		roomId, err := strconv.ParseInt(arr, 10, 64)

		if err != nil {
			log.Println("cannot parse room: ", err.Error())
			continue
		}

		realRoom, roomErr := api.GetRealRoom(roomId)

		if realRoom > 0 {
			roomSet.Add(realRoom)
		} else {
			log.Printf("房間 %v 無效，已略過 \n", roomId)
		}

		if roomErr != nil {
			_ = c.Error(roomErr)
			return nil, false
		}
	}

	// 有生之年我居然還要用 loop 轉泛型 array 同 type array, 醉了
	roomArr := roomSet.ToSlice()
	rooms := make([]int64, len(roomArr))

	for i, v := range roomArr {
		rooms[i] = v.(int64)
	}

	return rooms, true
}

func ActivateExpire(identifier string) {
	// 如果之前尚未有過訂閱 (即新增而不是更新)
	if _, subBefore := subscriber.Get(identifier); !subBefore {
		// 設置如果五分鐘後尚未連線 WebSocket 就清除訂閱記憶
		subscriber.ExpireAfter(identifier, time.After(time.Minute*5))
	}
}
