package main

import (
	"encoding/json"
	"strconv"
	"strings"
	//	"time"
	"flag"
	"fmt"
	"net"
	"os"
)

// Peers : Карта содержащая список пиров (дописать инфу о частях)
type Peers map[string]int

// Parts : Карта сожержащая начало области данных и номер части
type Parts map[int]int64

// Mssg : Структура для общения между нодами (сериализуется в json)
type Mssg struct {
	Header string `json:"header"`
	Part   int    `json:"part"`
	Data   []byte `json:"data"`
}

var (
	peerList Peers
	partSize int64 = 2097152 // 2 Mbytes in bytes
	partNums int             //= 0
	partList Parts
	sFile    *os.File
	fileSize int64  //= 0
	port     string = "8080"
)

func main() {
	var (
		iFile     = flag.String("if", "", "Путь до файла раздачи (для пида)")
		oFile     = flag.String("of", "", "Путь до файла раздачи (для сира)")
		targetAdr = flag.String("t", "", "Адрес для запроса c портом 8080")
	)

	flag.Parse()
	peerList = make(Peers)
	partList = make(Parts)

	fmt.Printf("MAIN: Init file\n") //DEBUG
	if *oFile != "" {
		tmpFile, err := os.OpenFile(*oFile, os.O_RDONLY, 0755)
		chkError(err)

		sFile = tmpFile
		defer sFile.Close()

		tmpLstat, err := os.Lstat(*oFile)
		chkError(err)

		fileSize = tmpLstat.Size()
		fmt.Printf("Размер файла раздачи- %d байт\n", fileSize) //DEBUG

		Mapper() // Обрабатывает данные файла

		fmt.Printf("Размер части - %d байт\n", partSize)           //DEBUG
		fmt.Printf("Кол-во частей - %v\n", partNums)               //DEBUG
		fmt.Printf("Список частей со смещениями - %d\n", partList) //DEBUG

	} else if *iFile != "" {
		tmpFile, err := os.OpenFile(*iFile, os.O_RDWR|os.O_CREATE, 0755) // Создаём файл с нуля
		chkError(err)

		sFile = tmpFile
		defer sFile.Close()
		partNums = 0
		partList[0] = int64(0)
	}

	fmt.Printf("Запускам слушателя\n") //DEBUG
	ln, err := net.Listen("tcp", ":"+port)
	chkError(err)
	defer ln.Close()

	if *targetAdr != "" {
		fmt.Printf("Режим пира\n")                         //DEBUG
		fmt.Printf("Коннект с сидом - %s\n\n", *targetAdr) //DEBUG
		go Loader(*targetAdr)
	} else {
		fmt.Printf("Режим сида\n\n") //DEBUG
		peerList[GetLocalIP()] = partNums
		fmt.Printf("Добавили себя, первого пира - %v\n", peerList) //DEBUG
	}

	for {
		conn, err := ln.Accept()
		chkError(err)
		s := conn.RemoteAddr().String()
		s = strings.Split(s, ":")[0]

		tmpPart := Reader(conn, sFile)

		if _, ok := peerList[s]; ok == false { // Если нет пира с таким ключом в мапе

			fmt.Printf("Добавлен пир: %s\n", s) //DEBUG
			peerList[s] = tmpPart               // Добавляем пира

			if *targetAdr == "" {
				var ta Mssg
				ta.Header = "PUSH"
				ta.Part = 0
				ta.Data = []byte(fmt.Sprintf("%v", peerList))
				b, err := json.Marshal(ta)
				chkError(err)

				fmt.Printf("Текущий список пиров, обновляем у пиров: %v\n", peerList) //DEBUG
				for addr, _ := range peerList {
					if addr != GetLocalIP() {
						Dialer(addr+":"+port, b)
					}
				}
			}
			fmt.Printf("\n") //DEBUG
		}
	}
}

// Mapper : Размечает файл на части
func Mapper() {
	if fileSize <= partSize {
		partNums = 0
		partList[0] = int64(0)
	} else {
		tmpStr := strconv.FormatInt(fileSize/partSize, 10)
		tmpNums, err := strconv.Atoi(tmpStr)
		chkError(err)

		partNums = tmpNums + 1 // +1 На случай недобора последней части до 2 Мбайт
		for i := 0; i < partNums; i++ {
			partList[i] = int64(i) * partSize
		}
	}
}

// GetLocalIP : Добывает свой локальный адрес по имени хоста
func GetLocalIP() string {
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			return fmt.Sprintf("%s", ipv4)
		}
	}
	return "localhost" //!!!!!!!
}

// Dialer : Производит отправку сообщений
func Dialer(ts string, msg []byte) {
	conn, err := net.Dial("tcp", ts)
	chkError(err)

	writedByte, err := conn.Write(msg)
	chkError(err)

	fmt.Printf("Байт отправлено: %d\n", writedByte) //DEBUG
	conn.Close()
}

func chkError(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

// Reader : Обрабатывает пришедшую инфу из соединения
func Reader(c net.Conn, f *os.File) int {
	var (
		a  Mssg // Входящее
		ta Mssg // Исходящее
	)

	s := c.RemoteAddr().String()
	tmpAddr := strings.Split(s, ":")[0]

	b := make([]byte, partSize)
	readedByte, err := c.Read(b)
	chkError(err)

	err = json.Unmarshal(b[0:readedByte], &a)
	chkError(err)

	switch a.Header {
	case "INITA":
		fmt.Printf("Получен запрос на инициализацию\n") //DEBUG

		ta.Header = "INITB"
		ta.Part = 0
		ta.Data = []byte(fmt.Sprintf("%d:%d", partNums, fileSize))

		tb, err := json.Marshal(ta)
		chkError(err)

		Dialer(tmpAddr+":"+port, tb)
		fmt.Printf("Отправили ответ на инициализацию\n") //DEBUG

	case "INITB":
		fmt.Printf("Пришел ответ на инициализацию: %s\n", s) //DEBUG

		pInfo := strings.Split(string(a.Data), ":")
		tmpNums, err := strconv.Atoi(pInfo[0])
		chkError(err)
		tmpSize, err := strconv.ParseInt(pInfo[1], 10, 64)
		chkError(err)

		partNums = tmpNums
		fileSize = tmpSize
		Mapper()

		fmt.Printf("Получено кл-во частей: %d\n", partNums)                //DEBUG
		fmt.Printf("Получен размер файла: %d\n", fileSize)                 //DEBUG
		fmt.Printf("Получен список частей со смещениями - %d\n", partList) //DEBUG

		//		case "GET":
		//			fmt.Printf("READER: GET part: %s\n", tmpArr[1])			//DEBUG
		//			tmpData := make([]byte, partSize)
		//
		//			count, err := f.ReadAt(tmpData, partList[tmpPart])
		//			if err != nil {
		//				if err.Error() != "EOF" {
		//					chkError(err)
		//				}
		//			}
		//
		//			data := string(tmpData[0:count])
		//
		//			fmt.Printf("READER: Send TAKE\n")				//DEBUG
		//			Dialer(tmpAddr + ":" + port, "TAKE:" + tmpArr[1] + ":" + data)

		//		case "TAKE":
		//
		//			tmpData := []byte(tmpArr[2])
		//			_, err := f.WriteAt(tmpData, partList[tmpPart])
		//			chkError(err)
		//
		//			fmt.Printf("READER: TAKE: %s\n", tmpArr[2])	//DEBUG

	case "PUSH":
		fmt.Printf("Пришло обновление списка пиров\n") //DEBUG
		tmpStr := string(a.Data)
		tmpStr = strings.Trim(tmpStr, "map[]")
		tmpArr := strings.Split(tmpStr, " ")
		for i := 0; i < len(tmpArr); i++ {
			pInfo := strings.Split(tmpArr[i], ":")
			if _, ok := peerList[pInfo[0]]; ok == false {
				tint, err := strconv.Atoi(pInfo[1])
				chkError(err)
				peerList[pInfo[0]] = tint
			}
		}
		fmt.Printf("Список пиров после обновления - %v\n", peerList) //DEBUG
	}
	fmt.Printf("\n") //DEBUG
	return a.Part
}

// Loader : Занимается сборкой файла
func Loader(ta string) {

	a := Mssg{
		Header: "INITA",
		Part:   0,
		Data:   []byte("0"),
	}

	b, err := json.Marshal(a)
	chkError(err)

	fmt.Printf("Отправлен запрос на инициализацию - %s\n", b) //DEBUG

	Dialer(ta, b)
	//	time.Sleep(1 * time.Second)
	//	for i, _ := range partList {
	//		Dialer(ta, "GET:" + strconv.Itoa(i))
	//		time.Sleep(1 * time.Second)
	//	}
}
