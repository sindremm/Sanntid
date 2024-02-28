package master_slave

import  (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"time"
)

const (
	SERVER_IP_ADDRESS = "127.0.0.1"
	PORT = "20005"
	FILENAME = "home/student/Documents/AjananMiaSindre/Sanntid/project/driver-go/master_slave/master_slave.go"
)

type Order struct {
    ID   int
    Item string
    // Add more fields as needed
}

type MasterSlave struct {
	current_order [4]*Order
	is_master	 bool
	backup_orders []*Order
	handle_order chan *Order
	switch_to_backup chan bool
}

func NewMasterSlave() *MasterSlave {
	return &MasterSlave{
		is_master: true,
		handle_order: make(chan *Order),
		switch_to_backup: make(chan bool),
	}
}

func (ms *MasterSlave) HandleOrderFromMaster(order *Order){
	if ms.is_master{
		ms.current_order = order
		fmt.Println("Master: Handling order ID", order.ID)
		//time.Sleep(3 * time.Second)
	} else {
		//slave handling order
		ms.backup_orders = append(ms.backup_orders, order)
		fmt.Println("Slave: Backup order ID", order.ID)
	}
}


func (ms *MasterSlave) SwitchToBackup(){
	ms.is_master = false
	fmt.Println("Master is dead, switching to backup")
}

var fullAddress = SERVER_IP_ADDRESS + ":" + PORT

func main() {
	start_time := time.Now()
	delay_milli := 100
	printCounter := time.Now()
	counter := 0

	ms := NewMasterSlave()

	filename := FILENAME

	for {
		if start_time.Add(time.Duration(delay_milli) * time.Millisecond).Before(time.Now()) {
			break
		}
		counter = listen(&start_time, counter, ms)
	}

	exec.Command("gnome-terminal", "--", "go", "run", filename).Run()

	for {
		send(counter)
		if printCounter.Add(time.Second).Before(time.Now()) {
			counter++
			fmt.Printf("%d\n", counter)
			printCounter = time.Now()
		}
	}
}

func send(counter int) {
	conn, err := net.Dial("udp", fullAddress)
	if err != nil {
		fmt.Printf("Some error 1 %v \n", err)
		return
	}
	defer conn.Close()
	
	_, err = fmt.Fprintf(conn, "%d", counter)
	if err != nil {
		fmt.Printf("error 3: %s \n", err)
	}
}

func listen(start_time *time.Time, counter int, ms *MasterSlave) (int){
	p := make([]byte, 2048)

	ServerAddr, err := net.ResolveUDPAddr("udp", fullAddress)
	if err != nil {
		fmt.Printf("Some error %v", err)
		return -1
	}

	conn, err := net.ListenUDP("udp", ServerAddr)
	if err != nil {
		fmt.Printf("Some error 1a %v", err)
		return -1
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	n, _, err := conn.ReadFromUDP(p)

	if err == nil {
		counter, err = strconv.Atoi(string(p[:n]))
		*start_time = time.Now()
	} 
	
	return counter
}