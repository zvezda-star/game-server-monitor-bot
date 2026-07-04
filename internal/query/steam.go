package query

import (
	"fmt"
	"net"
	"strings"
	"time"
)

func QuerySteam(ip string) (string, error) {
	if _, _, err := net.SplitHostPort(ip); err != nil {
		ip = ip + ":27015"
	}

	// пробую UDP
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

	query := []byte{
		0xFF, 0xFF, 0xFF, 0xFF,
		0x54, 0x53, 0x6F, 0x75, 0x72, 0x63, 0x65, 0x20,
		0x45, 0x6E, 0x67, 0x69, 0x6E, 0x65, 0x20, 0x51,
		0x75, 0x65, 0x72, 0x79, 0x00,
	}

	_, err = conn.Write(query)
	if err != nil {
		return querySteamTCP(ip)
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return querySteamTCP(ip)
	}

	if n < 5 {
		return querySteamTCP(ip)
	}

	if buf[4] != 0x49 {
		return querySteamTCP(ip)
	}

	// первый способ парсинга
	result, ok := parseSteamV1(buf[:n])
	if ok {
		return result, nil
	}

	// второй способ парсинга
	result, ok = parseSteamV2(buf[:n])
	if ok {
		return result, nil
	}

	return "Статус: Онлайн (подключение установлено)", nil
}

// первый способ: стандартный парсинг
func parseSteamV1(data []byte) (string, bool) {
	offset := 5
	if data[0] == 0xFF && data[1] == 0xFF && data[2] == 0xFF && data[3] == 0xFF {
		offset = 9
	}

	if offset >= len(data) {
		return "", false
	}

	// название сервера
	nameEnd := offset
	for nameEnd < len(data) && data[nameEnd] != 0 {
		nameEnd++
	}
	if nameEnd >= len(data) || nameEnd == offset {
		return "", false
	}
	serverName := string(data[offset:nameEnd])
	offset = nameEnd + 1

	// карта
	mapEnd := offset
	for mapEnd < len(data) && data[mapEnd] != 0 {
		mapEnd++
	}
	if mapEnd >= len(data) || mapEnd == offset {
		return "", false
	}
	mapName := string(data[offset:mapEnd])
	offset = mapEnd + 1

	// папка
	folderEnd := offset
	for folderEnd < len(data) && data[folderEnd] != 0 {
		folderEnd++
	}
	folder := string(data[offset:folderEnd])
	offset = folderEnd + 1

	// игра
	gameEnd := offset
	for gameEnd < len(data) && data[gameEnd] != 0 {
		gameEnd++
	}
	gameName := string(data[offset:gameEnd])
	offset = gameEnd + 1

	if offset+4 > len(data) {
		return "", false
	}
	players := int(data[offset])
	maxPlayers := int(data[offset+1])

	result := "Статус: Онлайн\n"
	if serverName != "" {
		result += fmt.Sprintf("Название: %s\n", serverName)
	}
	if gameName != "" {
		result += fmt.Sprintf("Игра: %s\n", gameName)
	} else {
		switch folder {
		case "cstrike":
			result += "Игра: Counter-Strike\n"
		case "tf":
			result += "Игра: Team Fortress 2\n"
		case "dayz":
			result += "Игра: DayZ\n"
		}
	}
	if mapName != "" {
		result += fmt.Sprintf("Карта: %s\n", mapName)
	}
	result += fmt.Sprintf("Игроки: %d / %d\n", players, maxPlayers)

	return result, true
}

// второй способ: через строки
func parseSteamV2(data []byte) (string, bool) {
	str := string(data)

	// ищу названия
	serverName := ""
	gameName := ""
	mapName := ""

	// первый кусок до первого нулевого байта
	parts := strings.SplitN(str, "\x00", 10)
	if len(parts) >= 1 && parts[0] != "" {
		serverName = parts[0]
	}
	if len(parts) >= 2 && parts[1] != "" {
		mapName = parts[1]
	}
	if len(parts) >= 4 && parts[3] != "" {
		gameName = parts[3]
	}

	if serverName == "" && gameName == "" {
		return "", false
	}

	result := "Статус: Онлайн\n"
	if serverName != "" {
		result += fmt.Sprintf("Название: %s\n", serverName)
	}
	if gameName != "" {
		result += fmt.Sprintf("Игра: %s\n", gameName)
	}
	if mapName != "" {
		result += fmt.Sprintf("Карта: %s\n", mapName)
	}
	result += "(информация получена частично)"

	return result, true
}

// резерв через TCP
func querySteamTCP(ip string) (string, error) {
	conn, err := net.DialTimeout("tcp", ip, 3*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()

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

	buf := make([]byte, 256)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		return "Статус: Онлайн (подключение установлено)", nil
	}

	if n > 0 {
		return "Статус: Онлайн (ответ получен по TCP)", nil
	}

	return "", fmt.Errorf("пустой ответ")
}