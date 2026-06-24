package query

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

func QueryMinecraft(ip string) (string, error) {
	// добавляю порт если не указан
	if _, _, err := net.SplitHostPort(ip); err != nil {
		ip = ip + ":25565"
	}

	// подключаюсь
	conn, err := net.DialTimeout("tcp", ip, 5*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// отправляю handshake
	host, port, _ := net.SplitHostPort(ip)
	handshake := []byte{
		0x00,                   // packet id
		0x47, 0x00, 0x00, 0x00, // protocol version 755
	}
	hostBytes := []byte(host)
	handshake = append(handshake, byte(len(hostBytes)))
	handshake = append(handshake, hostBytes...)
	p, _ := net.LookupPort("tcp", port)
	handshake = append(handshake, byte(p>>8), byte(p&0xFF))
	handshake = append(handshake, 0x01) // next state: status

	conn.Write(handshake)
	conn.Write([]byte{0x00}) // запрос статуса

	// читаю ответ
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}

	// ищу начало json
	jsonStart := -1
	for i := 1; i < n; i++ {
		if buf[i] == '{' {
			jsonStart = i
			break
		}
	}
	if jsonStart == -1 {
		return "", fmt.Errorf("не найден json в ответе")
	}

	// парсю json
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

	err = json.Unmarshal(buf[jsonStart:n], &status)
	if err != nil {
		return "", err
	}

	// формирую отчет
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

	return result, nil
}