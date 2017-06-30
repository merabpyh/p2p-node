package main

import (
	"strconv"
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

type Node struct {		// Структура локальной ноды
	Self		Peer
	Peers		Peers
	Con		net.Conn
}

func counter(a string) string {			// Набираем 100 символов в строку
	var str string = ""
	for i := 0; i < 100; i++ {
		str = str + a
	}
	return str
}

//func BytePart(p Peer) string {			// Временно генерим 1Кбайт инфы для тестовой отдачи XXXX
//	var a string = ""
//	switch p.PartNum {			// Раздаём 1012 символов
//		case "0":
//			a = counter("0")
//		case "1":
//			a = counter("1")
//		case "2":
//			a = counter("2")
//	}
//	return fmt.Sprintf("[%s]\n", a)
//}

func CheckError(err error) {            // Функция проверки ошибок
	if err != nil {
		log.Printf("Ошибка: %s\n", err)
	}
}

func NewNode(self Peer) *Node {         	// Функция инициализации экземпляра ноды
	n := new(Node)
	n.Self = self
	n.Peers = make(Peers)
	n.Peers[self.Address] = self
	n.Con = nil
	return n
}

func GetLocalIp() string {			// Определение локального ip адреса - возвращает строку с "IP"
	host, _ := os.Hostname()                                
	addrs, _ := net.LookupIP(host)                          
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			return fmt.Sprintf("%s", ipv4)
		}
	}
	return "localhost"
}

func GetPeerIP(c net.Conn) string {		// Функция вырезания адреса из соединения - возвращает строку с "IP"
	s := c.RemoteAddr().String()
	s = strings.Split(s, ":")[0]
	return s
}

func main() {					// MAIN()
	var (
		partNum string = "0"
		file = flag.String("f", "", "Путь до файла раздачи (для сида)")
		sid  = flag.String("s", "", "IP:Port раздающего сида (для пира)")
	)

	flag.Parse()

//	fmt.Printf("Main:Аргументы - %s\n", flag.Args())	//DEBUG
	if *file != "" {
		partNum, err := GenPartList(*file)		//!!!!!!!!
	}
	n := NewNode(Peer{GetLocalIp() + ":" + port, partNum})	// Создание экземпляра ноды
//	fmt.Printf("Main:Локальная нода -  %s\n", n.Self)	//DEBUG
	n.Waiter(*sid)
}

func (n *Node) Waiter(sid string) {				//!!!!!!!!!!!!Проверить тип!!!!!!!!!!!!!!!!!!!!!!!!!!!
	ln, err := net.Listen("tcp", ":" + port)                // Встаём на прослушку
	CheckError(err)

	for {
		c, err := ln.Accept()				// Ловим соединение
		n.Con = c
		CheckError(err)
		n.ParseRequest()				// Парсим запрос
	}
}

func (n *Node) ParseRequest() {
	b := make([]byte, 4096) 				// 4Kb Буфер
	bytesRead, err := n.Con.Read(b)				// Читаем байты из потока
	CheckError(err)

	tmpStr := string(b[0:bytesRead])			// Преобразуем в строку
	tmpArr := strings.Split(tmpStr, ":")			// Разделяем на части в массив

	switch tmpArr[0] {
	case "GIVEPART":							// [GIVEPART]:[PORT]:[PART]
		tmpPeer := Peer{GetPeerIP(n.Con) + ":" + tmpArr[1], tmpArr[2]}	// Получаем свежего пира
		if n.PeerAdd(tmpPeer) == true { n.PushPeers() }			// Добавляем пира в мапу
		go n.SendAnswer(tmpArr[2])					// Отвечаем пиру
	case "TAKEPART":							// [TAKEPART]:[PORT]:[PART]:[DATA]
		WritePart(tmpArr[2], tmpArr[3])					//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	case "PEERSUPD":							// [PEERSUPD]:[LIST]
		n.PeerListUpdate(tmpArr[2])					// Обновляем список пиров !!! проверить тип !!!
	}
}

func (n *Node) PeerAdd(p Peer) bool {				// Добавление пира в глобальный список ноды
	if tmpPeer, ok := n.Peers[p.Address]; ok == false {	// Если нет пира с таким ключом в мапе
		n.Peers[p.Address] = p				// Добавляем пира
		return true
	} else {
		if tmpPeer != p {				// Сравнение пира из локальной мапы с присланным пиром
			a, _ := strconv.Atoi(tmpPeer.PartNum)
			b, _ := strconv.Atoi(p.PartNum)
			if a < b {
				n.Peers[p.Address] = p		// Если значение части у присланного пира больше - перезаписываем на него
				return true
			}
		}
	}
	return false						// Возввращаем bool, добавили - true, не добавили - false	
//	fmt.Printf("PeerAdd:Актуальный список пиров\n")         //DEBUG
//	for i := range n.Peers {                                //DEBUG
//		fmt.Printf("%q\n", n.Peers[i])                  //DEBUG
//	}                                                       //DEBUG
}

func (n *Node) PeerListUpdate( l string) {		// Добавление пиров из присланного списка
	peerList := strings.Split( l, ":")		// Разделяем входную строку на части по символу ":"
	for key, value := range peerList {
		tmpStr := strings.Split(value, " ")
		n.PeerAdd(Peer{tmpStr[0], tmpStr[1]})	// Тут ловить ответ про добавление не нужно
	}
}

func (n *Node) PushPeers() {
	tmpStr := ""
	for addr, peer := range n.Peers {
		tmpStr = tmpStr + ":" + fmt.Sprint(peer)	// Добавляем в строчку каждого пира через ":"
	}
	b := []byte("PEERSUPD:" + tmpStr)		// Формируем буфер из строк
	for addr, peer := range n.Peers {
		c, err := net.Dial("tcp", peer.Address + ":" + port)    // Соединяемся c пиром из списка
		CheckError(err)

		bytesWrite, err := c.Write(b)                           // Суём ему буфер
		CheckError(err)

		fmt.Printf("PushPeers:Байт переданно -  %d\n", bytesWrite)      //DEBUG
		c.Close()                                               // Закрываем
	}
}

func (n *Node) SendAnswer(part string) {
	tmpb, err := ReadPart(part)				// Cчитываем нужную нам часть!!!!!!!!!
	CheckError(err)

	b := []byte("TAKEPART:" + part + ":")				// Формируем буфер для ответа

	bytesWrite, err := n.Con.Write(b + tmpb)		// Непосредственно запись в поток
	CheckError(err)

	fmt.Printf("SendAnswer:Байт переданно -  %d\n", bytesWrite)       //DEBUG
	n.Con.Close()
}


func (n *Node) GenPartList(file string) (partNums string, err error) {	// Генерируем карту смещений для файла раздачи
	
	f, err := os.Open(file)				// Открытие файла
	CheckError(err)

	tmpStat, err := f.Stat()
	fmt.Printf("GetPartList:Размер файла - %d байт\n",tmpStat.Size())
	f.Read(b []byte)
	f.Close()

	return partNums, err
}

func (n *Node) WritePart(part string, data string) {
	
}

func (n *Node) ReadPart(part string) []byte {
		
	return 
}
