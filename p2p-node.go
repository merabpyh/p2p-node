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
	peerList	Peers				// Список пиров
	partList	Parts				// Список частей файла
	port 		string 		= "8877"	// Порт по умоланию
	sFile		os.File				// Сам файл раздачи
)

type Part struct {				// Данные части файла
	Num		int			// Номер части
	Seek		int64			// Смещение до первого байта части
}

type Peer struct {				// Данные пира
	Address		string			// ip пира
	Num		int			// Номер последней запрошенной этим пиром части
	Con		net.Conn		// Соединение с этим пиром
}

type Peers map[string]Peer			// Список пиров ввиде мапы

type Parts map[int]Part				// Список файла ввиде мапы

func main() {					// MAIN()
	var (
		iFile = flag.String("if", "", "Путь до файла раздачи (для сида)")
		oFile = flag.String("of", "", "Путь до файла раздачи (для пира)")
		sid  = flag.String("s", "", "IP раздающего сида (для пира)")

	)

	flag.Parse()

	fmt.Printf("Main:Аргументы - %s\n", flag.Args())        //DEBUG

	if *sid != "" {
		fmt.Printf("Main: Выбрана роль - ПИР\n")	//DEBUG
		if *oFile != "" {
			sFile, err := os.OpenFile(*oFile, os.O_RDWR|os.O_CREATE, 0755)        // Создаём файл с нуля
			CheckError(err)

			tmpPeer := Peer{*sid, 0, nil}			// Формируем первого пира, он же сид

			peerList[tmpPeer.Address] = tmpPeer		// Суём сида в мампу

			c, err := net.Dial("tcp", *sid)			// Первый коннект всегда к сиду
			CheckError(err)

			bw := []byte("INIT")
			br := make([]byte, 10240)			// 10Kb Буфер
			bytesWrite, err := c.Write(bw)			// Непосредственно запись в поток
			bytesRead, err 	:= c.Read(br)			// Читаем байты из потока
			

//!!!!			функция слушанья для iGET(oTAKE) и (iPUSHPEERS)		// Слушаем и получаем  свежий список пиров 
										// Слушаем и получаем GET чтобы отдать TAKE
		} else {
			fmt.Printf("Main: Не указан выходной файл -of=\n")		//DEBUG
			os.Exit(1)
		}
	} else {
		fmt.Printf("Main: Выбрана роль - СИД\n")				//DEBUG
		if *iFile != "" {
			sFile, err := os.OpenFile(*oFile, os.O_RDONLY, 0755)
			CheckError(err)

//!!!!			partList, err := функция парсинга файла данных, для формирования списка частей

			tmpAdr := GetLocalIp()						// Получаем свой внешний адрес
			tmpPeer := Peer{tmpAdr, len(partList), nil}			// Формируем первого пира, это мы, мы и есть сид
			
			peerList[tmpPeer.Address]= tmpPeer				// Добавляем себя, сида, в список пиров

//!!!!			функция слушанья для iGET(oTAKE) запускает go (oPUSHPEERS)	// При обращении нового пира, мы добавляем его в список и рассылаем всем
											// Слушаем и получаем GET чтобы отдать TAKE
//			for {
				peerList.Worker()
//			}

		} else {
			fmt.Printf("Main: Не указан входной файл -if=\n")		//DEBUG
			os.Exit(1)
		}
	}
//	fmt.Printf("Main:Локальная нода -  %s\n", GetLocalIp() + ":" + port)		//DEBUG
//      os.Exit(0)
}

func CheckError(err error) {            // Функция проверки ошибок
	if err != nil {
		log.Printf("Ошибка: %s\n", err)
	}
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

func (pl *Peers) Worker() {					//
	ln, err := net.Listen("tcp", ":" + port)                // Встаём на прослушку
	CheckError(err)
	defer ln.Close()

//	Getter()

	for {
		c, err := ln.Accept()				// Ловим соединение // Получаем запрос // Отправляем запрос
		CheckError(err)

		tmpAdr := GetPeerIP(c)				// Получаем адрес из коннекта
		tmpPeer := Peer{tmpAdr, 0, c}			// Формируем пира по инфе из коннекта

		ok := PeerCheck(tmpPeer)			// Проверяем знаком ли пир
		if ok == true {
			pl.PushPeers()				// Пушим обновлённый список пиров всем пирам
		}
		
	}
}

func Getter() {
	
}

func Parser() {
	b := make([]byte, 10240) 				// 10Kb Буфер
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

func  PeerCheck(p Peer) bool {			// Проверка на наличие пира в мапе
	if tmpPeer, ok := peerList[p.Address]; ok == false {	// Если нет пира с таким ключом в мапе
		peerList[p.Address] = p				// Добавляем пира
		
		return true
	} else {
		if tmpPeer != p {				// Сравнение пира выдернутого из локальной мапы с присланным пиром
			a, _ := strconv.Atoi(tmpPeer.PartNum)
			b, _ := strconv.Atoi(p.PartNum)
			if a < b {
				peerList[p.Address] = p		// Если значение части у присланного пира больше - перезаписываем на него
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

func (pl *Peers) PushPeers() {
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


func GenPartList(file string) (int64, *os.File, error) {	// Генерируем карту смещений для файла раздачи
	var partNums int64 = 0
	f, err := os.Open(file)				// Открытие файла
	CheckError(err)

	tmpLstat, err := os.Lstat(file)
	CheckError(err)

//	tmpStat, err := f.Stat()
//	CheckError(err)
//	size := tmpStat.Size()
	size := tmpLstat.Size()
	fmt.Printf("GetPartList:Размер файла - %d байт\n", size)
	partNums = size /10

	if size >= 10000 {
		partNums = size / 1000
	} else {
		if size >= 1000 {
			partNums = size / 100
		} else {
			if size >= 100 {
				partNums = size / 10
			}
		}
	}	
	
//	f.Close()

	return partNums, f, err
}
