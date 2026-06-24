package query

import (
	"fmt"
	"net"
	"time"
)

// проверяю статус стим сервера
func QuerySteam(ip string) (string, error) {
	// добавляю порт если не указан
	host, port, err := net.SplitHostPort(ip)
	if err != nil {
		host = ip
		port = "27015"
		ip = host + ":" + port
	}

	// подключаюсь по udp
	addr, err := net.ResolveUDPAddr("udp", ip)
	if err != nil {
		return "", err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// формирую a2s_info запрос
	query := []byte{
		0xFF, 0xFF, 0xFF, 0xFF,
		0x54, 0x53, 0x6F, 0x75, 0x72, 0x63, 0x65, 0x20,
		0x45, 0x6E, 0x67, 0x69, 0x6E, 0x65, 0x20, 0x51,
		0x75, 0x65, 0x72, 0x79, 0x00,
	}

	_, err = conn.Write(query)
	if err != nil {
		return "", err
	}

	// читаю ответ
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}

	if n < 5 || buf[4] != 0x49 {
		return "", fmt.Errorf("неверный ответ от сервера")
	}

	offset := 5
	if buf[0] == 0xFF && buf[1] == 0xFF && buf[2] == 0xFF && buf[3] == 0xFF {
		offset = 9
	}

	// читаю название сервера
	nameEnd := offset
	for nameEnd < n && buf[nameEnd] != 0 {
		nameEnd++
	}
	if nameEnd >= n {
		return "", fmt.Errorf("неверное название сервера")
	}
	serverName := string(buf[offset:nameEnd])
	offset = nameEnd + 1

	// читаю карту
	mapEnd := offset
	for mapEnd < n && buf[mapEnd] != 0 {
		mapEnd++
	}
	if mapEnd >= n {
		return "", fmt.Errorf("неверное название карты")
	}
	mapName := string(buf[offset:mapEnd])
	offset = mapEnd + 1

	// пропускаю папку
	folderEnd := offset
	for folderEnd < n && buf[folderEnd] != 0 {
		folderEnd++
	}
	folder := string(buf[offset:folderEnd])
	offset = folderEnd + 1

	// читаю название игры
	gameEnd := offset
	for gameEnd < n && buf[gameEnd] != 0 {
		gameEnd++
	}
	gameName := string(buf[offset:gameEnd])
	offset = gameEnd + 1

	if offset+4 > n {
		return "", fmt.Errorf("неверные данные об игроках")
	}
	players := int(buf[offset])
	maxPlayers := int(buf[offset+1])
	offset += 2

	offset += 2

	osType := buf[offset]
	offset++

	offset += 2

	versionEnd := offset
	for versionEnd < n && buf[versionEnd] != 0 {
		versionEnd++
	}
	version := string(buf[offset:versionEnd])

	// определяю игру по папке если название не пришло
	if gameName == "" {
		switch folder {
		case "cstrike":
			gameName = "Counter-Strike"
		case "tf":
			gameName = "Team Fortress 2"
		case "dayz":
			gameName = "DayZ"
		case "left4dead2":
			gameName = "Left 4 Dead 2"
		case "garrysmod":
			gameName = "Garry's Mod"
		case "insurgency":
			gameName = "Insurgency"
		default:
			gameName = folder
		}
	}

	// формирую отчет
	result := "Статус: Онлайн\n"
	result += fmt.Sprintf("Название: %s\n", serverName)
	result += fmt.Sprintf("Игра: %s\n", gameName)
	result += fmt.Sprintf("Карта: %s\n", mapName)
	result += fmt.Sprintf("Игроки: %d / %d\n", players, maxPlayers)

	if osType == 'w' {
		result += "ОС: Windows\n"
	} else if osType == 'l' {
		result += "ОС: Linux\n"
	}

	if version != "" {
		result += fmt.Sprintf("Версия: %s\n", version)
	}

	return result, nil
}