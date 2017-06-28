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

type Node struct {		// Структура локальной ноды
	Self		Peer
	Peers		Peers
	peerCheck	bool
}

//func counter(a string) string {			// Набираем 1012 символов XXX
//	var str string = ""
//	for i := 0; i < 1012; i++ {
//		str = str + a
//	}
//	return str
//}

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
	str := c.RemoteAddr().String()
	str = strings.Split(str, ":")[0]
	return str
}

func main() {					// MAIN()
	file := flag.String("f", "", "Путь до файла раздачи (для сида)")
	sid  := flag.String("s", "", "IP:Port раздающего сида (для пира)")
	flag.Parse()
//	fmt.Printf("Main:Аргументы - %s\n", flag.Args())	//DEBUG
	if file != "" {
		partNum, err := GenPartList(*file)			//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	} else {
		partNum := 0
	}
	CheckError(err)
	n := NewNode(Peer{GetLocalIp() + ":" + port, partNum})	// Создание экземпляра ноды
//	fmt.Printf("Main:Локальная нода -  %s\n", n.Self)	//DEBUG
	n.Waiter(*sid)
}

func (n *Node) Waiter(sid *string) {				//!!!!!!!!!!!!Проверить тип!!!!!!!!!!!!!!!!!!!!!!!!!!!
	ln, err := net.Listen("tcp", ":" + port)                // Встаём на прослушку
	CheckError(err)
	
	for {
		c, err := ln.Accept()				// Ловим соединение
		CheckError(err)
		n.ParseRequest(c)				// Парсим запрос
	}
}

func (n *Node) ParseRequest(c *net.Conn) {
	b := make([]byte, 4096) 				// 4Kb Буфер
	bytesRead, err := c.Read(b)				// Читаем байты из потока
	CheckError(err)

	tmpStr := string(b[0:bytesRead])			// Преобразуем в строку
	tmpArr := strings.Split(tmpStr, ":")			// Разделяем на части в массив

	switch tmpArr[0] {
	case "GIVEPART":							// [GIVEPART]:[PORT]:[PART]
		tmpPeer := Peer{GetPeerIP(c) + ":" + tmpArr[1], tmpArr[2]}	// Получаем свежего пира
		n.PeerAdd(tmpPeer)						// Добавляем в список
		go SendAnswer(c, tmpArr[2])					// Отвечаем пиру
	case "TAKEPART":							// [TAKEPART]:[PORT]:[PART]:[DATA]
		WritePart(tmpArr[2], tmpArr[3])					//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	case "PEERSUPD":							// [PEERSUPD]:[PORT]:[LIST]
		n.PeerListUpdate(tmpArr[2])					//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	}
}

func (n *Node) PeerAdd(p Peer) {                                // Добавление пира в глобальный список ноды
	_, ok := n.Peers[p.Address]                             // Ищем в мапе по адресу (ключу) пира
	n.Peers[p.Address] = p
	if ok != true {                                         // Не нашли
		PushPeers()					//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	}
	
	fmt.Printf("PeerAdd:Актуальный список пиров\n")         //DEBUG
	for i := range n.Peers {                                //DEBUG
		fmt.Printf("%q\n", n.Peers[i])                  //DEBUG
	}                                                       //DEBUG
}

func (n *Node) PeerListUpdate(list map[string]Peer) {
	
}

func PushPeers() {
	
}

func SendAnswer(c *net.Conn, part string) {
	tmpb, err := ReadPart(part)				// Cчитываем нужную нам часть
	CheckError(err)

	b := []byte("TAKEPART:" + p.PartNum + ":")		// Формируем буфер для ответа

	bytesWrite, err := c.Write(b + tmpb)			// Непосредственно запись в поток
	CheckError(err)

	fmt.Printf("SendAnswer:Байт переданно -  %d\n", bytesWrite)       //DEBUG
	c.Close()
}

func GenPartList(file string) {
	
}

func WritePart(part string) {
	
}

func ReadPart(part string) {
	
}
