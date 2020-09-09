package supervisor

import (
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	"sync"
	"time"

	"go.uber.org/zap"
)

var SupervisorChannel chan model.OrderStatus

var KitchenChannel chan model.Order = nil

var DispatchChannel chan model.Order = nil

var StorageChannel chan model.ShelfItem = nil

var NewSpaceAvailableChannel chan string = nil

var OverflownChannel chan model.ShelfItem = nil

var DispatchWaitGroup sync.WaitGroup

var Report *ReportBook

var lastActivityReportedTime time.Time = time.Now()
var lastActivityHealthCheckedTime time.Time = time.Now()

type ReportBook struct {
	index  map[string]model.OrderStatus
	status map[string][]model.OrderStatus
	locker sync.Mutex
}

func (r *ReportBook) IsTrashed(orderId string) bool {
	r.locker.Lock()
	defer r.locker.Unlock()

	// Get last known status
	status, isPresent := r.index[orderId]
	return isPresent && status.Status == model.ORDER_EXPIRED
}

func (r *ReportBook) IsEvicted(orderId string) bool {
	r.locker.Lock()
	defer r.locker.Unlock()

	// Get last known status
	order, isPresent := r.index[orderId]
	return isPresent && order.Status == model.ORDER_EVICTED
}

func (r *ReportBook) push(order model.OrderStatus) {
	r.locker.Lock()
	defer r.locker.Unlock()

	// Maintain last known status of an order
	r.index[order.OrderId] = order

	// Maintain record of orders by status
	if r.status[order.Status] == nil {
		r.status[order.Status] = make([]model.OrderStatus, 0)
	}
	r.status[order.Status] = append(r.status[order.Status], order)
}

func (r *ReportBook) GenerateReport() {
	r.locker.Lock()
	defer r.locker.Unlock()

	var totalOrdersReceived float32 = 0
	var totalOrdersProcessed float32 = 0
	var totalOrdersExpired float32 = 0
	var totalOrdersEvicted float32 = 0
	var totalOrdersPickedUp float32 = 00

	orders, isPresent := r.status[model.ORDER_RECEIVED]
	if isPresent {
		totalOrdersReceived = float32(len(orders))
	}

	orders, isPresent = r.status[model.ORDER_PICKED]
	if isPresent {
		totalOrdersPickedUp = float32(len(orders))
	}

	orders, isPresent = r.status[model.ORDER_PROCESSED]
	if isPresent {
		totalOrdersProcessed = float32(len(orders))
	}

	orders, isPresent = r.status[model.ORDER_EXPIRED]
	if isPresent {
		totalOrdersExpired = float32(len(orders))
	}

	orders, isPresent = r.status[model.ORDER_EVICTED]
	if isPresent {
		totalOrdersEvicted = float32(len(orders))
	}

	// Print report
	zap.S().Infof("===============Order Status Report===============")
	zap.S().Infof("Total Orders Received: %.0f", totalOrdersReceived)
	zap.S().Infof("Total Orders Processed: %.0f", totalOrdersProcessed)
	zap.S().Infof("Total Orders Picked-Up: %.0f", totalOrdersPickedUp)
	zap.S().Infof("Total Orders Expired: %.0f", totalOrdersExpired)
	zap.S().Infof("Total Orders Evicted: %.0f", totalOrdersEvicted)

	zap.S().Infof("Orders processed percentage: %.2f%% ", (totalOrdersProcessed/totalOrdersReceived)*100)
	zap.S().Infof("Orders delivery percentage: %.2f%% ", (totalOrdersPickedUp/totalOrdersProcessed)*100)
	zap.S().Infof("Orders expired percentage: %.2f%%", (totalOrdersExpired/totalOrdersProcessed)*100)
	zap.S().Infof("Orders evicted percentage: %.2f%%", (totalOrdersEvicted/totalOrdersProcessed)*100)

	zap.S().Infof("Overall Orders success percentage: %.2f%%", (totalOrdersPickedUp/totalOrdersReceived)*100)
	zap.S().Infof("===============End Report===============")
}

func Start(noOfOrdersToRead int) {
	SupervisorChannel = make(chan model.OrderStatus, noOfOrdersToRead)

	KitchenChannel = make(chan model.Order, noOfOrdersToRead)
	DispatchChannel = make(chan model.Order, noOfOrdersToRead)
	StorageChannel = make(chan model.ShelfItem, noOfOrdersToRead)

	NewSpaceAvailableChannel = make(chan string, noOfOrdersToRead)
	OverflownChannel = make(chan model.ShelfItem, noOfOrdersToRead)

	Report = &ReportBook{
		index:  make(map[string]model.OrderStatus),
		status: make(map[string][]model.OrderStatus),
	}
	process()
}

func process() {
	go func() {
		for {
			select {
			case reportMsg, isOpen := <-SupervisorChannel:
				if !isOpen {
					SupervisorChannel = nil
					break
				}

				Report.push(reportMsg)
				zap.S().Infof("Supervisor: Order '%s' is reported to supervisor with status %s", reportMsg.OrderId, reportMsg.Status)
				lastActivityReportedTime = time.Now()
			default:
				handleNoMsgReceived()
			}

			if SupervisorChannel == nil {
				break
			}
		}
	}()
}

func handleNoMsgReceived() {
	idealTimeS := 10.0
	now := time.Now()
	if now.Sub(lastActivityReportedTime).Seconds() >= idealTimeS && now.Sub(lastActivityHealthCheckedTime).Seconds() >= idealTimeS {
		zap.S().Infof("---Supervisor: No orders received for more than %.0f(s). Press 'Ctrl + C' to terminate---", idealTimeS)
		lastActivityHealthCheckedTime = now
	}
}

func CloseAll() {
	close(SupervisorChannel)
	close(KitchenChannel)
	close(DispatchChannel)
	close(StorageChannel)
	close(NewSpaceAvailableChannel)
	close(OverflownChannel)
	Report = nil
}
