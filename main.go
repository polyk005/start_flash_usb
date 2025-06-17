package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"time"
)

type AppConfig struct {
	Name string
	Path string
	Args []string
}

func main() {
	clearConsole()
	setGreenText()
	showDedSecArt()
	time.Sleep(5 * time.Second)

	fmt.Println("\n[!] Взлом системы запущен...")
	fakeHackAnimation()

	apps := []AppConfig{
		{Name: "Chrome", Path: "./apps/chrome.exe", Args: []string{"/silent"}},
		{Name: "7-Zip", Path: "./apps/7z.exe", Args: []string{"/S"}},
	}

	for _, app := range apps {
		installApp(app)
	}

	fmt.Println("\n✅ Установка завершена. DedSec гордится вами!")
	pause()
}

func showDedSecArt() {
	fmt.Println(`
  ██████╗ ███████╗██████╗ ███████╗
 ██╔═══██╗██╔════╝██╔══██╗██╔════╝
 ██║   ██║█████╗  ██║  ██║███████╗
 ██║   ██║██╔══╝  ██║  ██║╚════██║
 ╚██████╔╝██║     ██████╔╝███████║
  ╚═════╝ ╚═╝     ╚═════╝ ╚══════╝
`)
}

func fakeHackAnimation() {
	phases := []struct {
		name         string
		delay        time.Duration
		glitchChance int
	}{
		{"Сканирование сети", 800, 20},
		{"Обход межсетевого экрана", 500, 30},
		{"Подбор учетных данных", 700, 40},
		{"Эскалация привилегий", 600, 25},
	}

	targets := []string{
		"Сервер CTOS v2.3.5",
		"Банк 'Pacific' (FIB#3341)",
		"Умный район Blume",
		"Трафик камер ALX-9",
	}

	// Начало атаки
	fmt.Println("\n\x1b[34m[+] Инициализация руткита DedSec_v9...\x1b[0m")
	time.Sleep(1 * time.Second)

	// Фазы взлома
	for phaseNum, phase := range phases {
		fmt.Printf("\n\x1b[36m[%d/%d] %s...\x1b[0m\n", phaseNum+1, len(phases), phase.name)

		for i := 0; i < 100; {
			// Случайный прогресс
			step := rand.Intn(15) + 5
			if i+step > 100 {
				i = 100
			} else {
				i += step
			}

			// Глюки системы
			if rand.Intn(100) < phase.glitchChance {
				glitchTypes := []string{
					"TRACE DETECTED",
					"SIGNATURE VERIFICATION FAILED",
					"CONNECTION RESET",
					"ROOTKIT ALERT",
				}
				fmt.Printf("\r\x1b[31m[!] %s\x1b[0m%-40s",
					glitchTypes[rand.Intn(len(glitchTypes))], "")
				time.Sleep(time.Duration(200+rand.Intn(300)) * time.Millisecond)
				fmt.Printf("\r%-60s\r", "")
				continue
			}

			// Основная анимация
			target := targets[rand.Intn(len(targets))]
			ip := fmt.Sprintf("%d.%d.%d.%d:%d",
				rand.Intn(255), rand.Intn(255),
				rand.Intn(255), rand.Intn(255),
				rand.Intn(65535))

			// Стилизованный вывод
			fmt.Printf("\r>> %-25s [%-20s] \x1b[33m%3d%%\x1b[0m",
				truncate(target, 25),
				ip,
				i)

			// Динамическая задержка
			time.Sleep(time.Duration(phase.delay+time.Duration(rand.Intn(2))) * time.Millisecond)
		}

		// Финальный статус фазы
		fmt.Printf("\r\x1b[32m[+] %s УСПЕШНО\x1b[0m%-30s\n",
			phase.name, "")
		time.Sleep(500 * time.Millisecond)
	}

	// Финальный взлом
	fmt.Println("\n\x1b[5;32m[!] СИСТЕМА СКОМПРОМЕТИРОВАНА\x1b[0m")
	time.Sleep(1 * time.Second)
	fmt.Println("\x1b[32m[+] Установка backdoor...\x1b[0m")
	time.Sleep(2 * time.Second)
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

func installApp(app AppConfig) {
	fmt.Printf("\n🔧 Устанавливаем %s...\n", app.Name)

	cmd := exec.Command(app.Path, app.Args...)
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Ошибка: %v\n", err)
	} else {
		fmt.Printf("✔ %s взломан и установлен!\n", app.Name)
	}
}

func clearConsole() {
	fmt.Print("\033[H\033[2J")
}

func setGreenText() {
	fmt.Print("\033[32m")
}

func pause() {
	fmt.Print("\nНажмите Enter...")
	fmt.Scanln()
}
