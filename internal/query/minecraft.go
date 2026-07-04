package query

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

func QueryMinecraft(ip string) (string, error) {
	if _, _, err := net.SplitHostPort(ip); err != nil {
		ip = ip + ":25565"
	}

	conn, err := net.DialTimeout("tcp", ip, 5*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	host, port, _ := net.SplitHostPort(ip)
	handshake := []byte{
		0x00, 0x47, 0x00, 0x00, 0x00,
	}
	hostBytes := []byte(host)
	handshake = append(handshake, byte(len(hostBytes)))
	handshake = append(handshake, hostBytes...)
	p, _ := net.LookupPort("tcp", port)
	handshake = append(handshake, byte(p>>8), byte(p&0xFF))
	handshake = append(handshake, 0x01)

	conn.Write(handshake)
	conn.Write([]byte{0x00})

	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}

	// пробую первый способ парсинга
	result, ok := parseMinecraftV1(buf[:n])
	if ok {
		return result, nil
	}

	// пробую второй способ парсинга
	result, ok = parseMinecraftV2(buf[:n])
	if ok {
		return result, nil
	}

	// пробую третий способ парсинга
	result, ok = parseMinecraftV3(buf[:n])
	if ok {
		return result, nil
	}

	// если ничего не получилось — просто говорю что онлайн
	return "Статус: Онлайн (подключение установлено)", nil
}

// первый способ: стандартный JSON
func parseMinecraftV1(data []byte) (string, bool) {
	var status struct {
		Version struct {
			Name string `json:"name"`
		} `json:"version"`
		Players struct {
			Online int `json:"online"`
			Max    int `json:"max"`
			Sample []struct {
				Name string `json:"name"`
			} `json:"sample"`
		} `json:"players"`
		Description struct {
			Text string `json:"text"`
		} `json:"description"`
	}

	jsonStart := -1
	for i := 1; i < len(data); i++ {
		if data[i] == '{' {
			jsonStart = i
			break
		}
	}
	if jsonStart == -1 {
		return "", false
	}

	err := json.Unmarshal(data[jsonStart:], &status)
	if err != nil {
		return "", false
	}

	result := "Статус: Онлайн\n"
	if status.Description.Text != "" {
		result += fmt.Sprintf("Название: %s\n", status.Description.Text)
	}
	if status.Version.Name != "" {
		result += fmt.Sprintf("Версия: %s\n", status.Version.Name)
	}
	result += fmt.Sprintf("Игроки: %d / %d\n", status.Players.Online, status.Players.Max)

	if len(status.Players.Sample) > 0 {
		result += "Список игроков:\n"
		for _, p := range status.Players.Sample {
			if p.Name != "" {
				result += fmt.Sprintf("- %s\n", p.Name)
			}
		}
	}
	return result, true
}

// второй способ: через строки (вырезаю нужные поля)
func parseMinecraftV2(data []byte) (string, bool) {
	str := string(data)

	// ищу "players"
	playersIdx := strings.Index(str, `"players"`)
	if playersIdx == -1 {
		return "", false
	}

	// ищу "online"
	onlineIdx := strings.Index(str[playersIdx:], `"online"`)
	if onlineIdx == -1 {
		return "", false
	}

	// ищу "max"
	maxIdx := strings.Index(str[playersIdx:], `"max"`)
	if maxIdx == -1 {
		return "", false
	}

	// вырезаю числа
	var online, max int
	fmt.Sscanf(str[playersIdx+onlineIdx:], `"online":%d`, &online)
	fmt.Sscanf(str[playersIdx+maxIdx:], `"max":%d`, &max)

	result := "Статус: Онлайн\n"
	if online > 0 || max > 0 {
		result += fmt.Sprintf("Игроки: %d / %d\n", online, max)
	}

	// ищу название сервера
	descIdx := strings.Index(str, `"description"`)
	if descIdx != -1 {
		textIdx := strings.Index(str[descIdx:], `"text"`)
		if textIdx != -1 {
			start := strings.Index(str[descIdx+textIdx:], `"`) + descIdx + textIdx + 1
			end := strings.Index(str[start:], `"`)
			if end != -1 && start+end < len(str) {
				name := str[start : start+end]
				if name != "" {
					result += fmt.Sprintf("Название: %s\n", name)
				}
			}
		}
	}

	return result, true
}

// третий способ: тупо ищу цифры
func parseMinecraftV3(data []byte) (string, bool) {
	str := string(data)

	// ищу любые цифры похожие на игроков
	var online, max int
	found := false

	// ищу "players"
	playersIdx := strings.Index(str, `"players"`)
	if playersIdx != -1 {
		// ищу две цифры подряд через запятую
		part := str[playersIdx:]
		// ищу "online"
		oIdx := strings.Index(part, `"online"`)
		if oIdx != -1 {
			fmt.Sscanf(part[oIdx:], `"online":%d`, &online)
			found = true
		}
		// ищу "max"
		mIdx := strings.Index(part, `"max"`)
		if mIdx != -1 {
			fmt.Sscanf(part[mIdx:], `"max":%d`, &max)
			found = true
		}
	}

	if !found {
		return "", false
	}

	result := "Статус: Онлайн\n"
	if online > 0 || max > 0 {
		result += fmt.Sprintf("Игроки: %d / %d\n", online, max)
	}
	result += "(информация получена частично)"

	return result, true
}