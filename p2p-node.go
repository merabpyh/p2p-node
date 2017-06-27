package main

import (
	"strings"
//	"bytes"
	"flag"
	"net"
	"fmt"
	"log"
	"os"
)

var (
	port string = "8877"	// Порт по умоланию
)

type Peer struct {		// Структура хранит данные пира
	Address		string	// ip:port пира
	PartNum	 	string	// номер последней запрошенной части
}

type Peers map[string]Peer	// Список пиров ввиде карты

type Node struct {			// Структура локальной ноды
	Self		Peer
	Peers		Peers

	peerCheck	bool
}

func NewNode(self Peer) *Node {		// Функция инициализации экземпляра ноды
	n := new(Node)
	n.Self = self
	n.Peers = make(Peers)
	n.Peers[self.Address] = self
	return n
}

func CheckError(err error) {		// Функция проверки ошибок
	if err != nil {
		log.Printf("Ошибка: %s\n", err)
	}
}

func GetPeerIP(c net.Conn) string {				// Функция вырезания адреса из соединения - возвращает строку с "IP"
	str := c.RemoteAddr().String()
	str = strings.Split(str, ":")[0]
	return str
}

func GetLocalIp() string {					// Определение локального ip адреса - возвращает строку с "IP"
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			return fmt.Sprintf("%s", ipv4)
		}
	}
	return "localhost"
}

func RequestParse(c net.Conn) (pTmp Peer, pStat bool) {		// Функция парсит запрос -  возвращает Peer

	b := make([]byte, 4096)
	bytesRead, err := c.Read(b)				// Читаем байты из потока
	CheckError(err)

	tmpStr := string(b[0:bytesRead])			// Преобразуем в строку
	fmt.Printf("RequestParse:Полученные данные\n")		//DEBUG
	fmt.Printf("%s\n", tmpStr)				//DEBUG

	tmpArr := strings.Split(tmpStr, ":")			// (0)[GIVEPART]:(1)[part_num]:(2)[PORT]

		if tmpArr[0] == "GIVEPART" {			// ok
			pTmp = Peer{GetPeerIP(c) + ":" + tmpArr[2], tmpArr[1]}
			pStat = false
		} else {					// not ok
			pTmp = Peer{"XXX", "XXX"}
			pStat = true
		}
	
	return pTmp, pStat
}

func (n *Node) PeerAdd(p Peer) {				// Добавление пира в глобальный список ноды
	_, ok := n.Peers[p.Address]				// Ищем в мапе по адресу (ключу) пира
	fmt.Printf("PeerAdd:Ищем пира в списке -  %t\n", ok)	//DEBUG

	if ok != true {						// Не нашли
		fmt.Printf("PeerAdd:Добавляем пира\n")		//DEBUG
		n.peerCheck = true
//		n.Peers[p.Address] = p
	} else {						// Нашли
		fmt.Printf("PeerAdd:Пир в списке - обновляем части\n")	//DEBUG
//		n.Peers[p.Address] = p
	}
	
	n.Peers[p.Address] = p

	fmt.Printf("PeerAdd:Актуальный список пиров\n")         //DEBUG
	for i := range n.Peers {                                //DEBUG
	        fmt.Printf("%q\n", n.Peers[i])                  //DEBUG
	}                                                       //DEBUG

}

func SendPart(c net.Conn, p Peer) {					// Отправка данных обратно пиру

	b := []byte("TAKEPART:" + p.PartNum + ":" + BytePart(p))	// Формируем буфер для ответа, считывая нужную нам часть

	bytesWrite, err := c.Write(b)					// Непосредственно запись в поток
	CheckError(err)

        fmt.Printf("SendPart:Байт переданно -  %d\n", bytesWrite)	//DEBUG
	c.Close()							// Отдали данные части, закрыли коннект и забыли
}

func counter(a string) string {			// Набираем 1012 символов
	var str string = ""
	for i := 0; i < 1012; i++ {
		str = str + a
	}
	return str
}

func BytePart(p Peer) string {			// ТЕСТОВАЯ ПРОСЛОЙКА	// Временно генерим 1Кбайт инфы для тестовой отдачи
	var a string = ""

	switch p.PartNum {			// Раздаём 1012 символов
		case "0":
			a = counter("0")
		case "1":
			a = counter("1")
		case "2":
			a = counter("2")
	}

	return fmt.Sprintf("[%s]\n", a)
}

func main() {							//

	role := flag.Bool("r", true, "Роль ноды: сид - 1, пир - 0")
//	file := flag.String("f", "", "Путь до файла раздачи (для сида)")
	sid  := flag.String("s", "", "IP:Port раздающего сида (для пира)")
	flag.Parse()

	fmt.Printf("Main:Аргументы - %s\n", flag.Args())	//DEBUG

	n := NewNode(Peer{GetLocalIp() + ":" + port, "Всё"})
	fmt.Printf("Main:Локальная нода -  %s\n", n.Self)	//DEBUG

	if *role == true {
		fmt.Printf("Main:Роль раздающего\n")		//DEBUG
		n.seeder()
	} else {
		fmt.Printf("Main:Роль качающего\n")		//DEBUG
		n.peerer(*sid)
	}
}

func (n *Node) seeder() {					// Поведение сида

	ln, err := net.Listen("tcp", ":" + port)		// Встаём на прослушку
	CheckError(err)

	for {
		c, err := ln.Accept()				// Приём соединения
		CheckError(err)
		
		tmpPeer, errBool := RequestParse(c) 		// Чтение и обработка из потока - возвращает пира {адрес, часть}
		if errBool == false {
			n.PeerAdd(tmpPeer)			// Добавдение пира в список
		}

		fmt.Printf("Seeder:Отправляем данные пиру - %s\n", tmpPeer)     //DEBUG - Вывод сообщения о внесении в пиры

		go SendPart(c, tmpPeer)				// Отправка данных пиру
		
		
	}
}

func (n *Node) peerer(ip string) {				// Поведение пира

	b := []byte("GIVEPART:1:8877")				// Тестовый запрос
	d := make([]byte, 4096) 

	conn, err := net.Dial("tcp", ip)			// Установка соединения
	CheckError(err)

	bytesWrite, err := conn.Write(b)			// Отправка инфы
	CheckError(err)

	fmt.Printf("Peerer:Байт переданно - %d\n", bytesWrite)	//DEBUG

	bytesRead, err := conn.Read(d)				// Читаем ответ - не универсально
	CheckError(err)

	fmt.Printf("Peerer:Байт получено - %d\n", bytesRead)	//DEBUG

	tmpStr := string(d[0:bytesRead])			// Преобразуем в строку то что получили - не универсально
	
	fmt.Printf("%s\n", tmpStr)				//DEBUG

//	os.Exit(0)
// тут тоже кучу всего менять
}
