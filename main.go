package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"sync"
	"time"
)

type AppConfig struct {
	Name     string
	Path     string
	Args     []string
	CheckCmd string
	CheckPath string
}

var (
	installMutex  sync.Mutex
	installedApps = make(map[string]bool)
	successCount  int
	failedCount   int
	totalDuration time.Duration
)

func main() {
	clearConsole()
	setGreenText()
	showDedSecArt()
	time.Sleep(3 * time.Second)

	fmt.Println("\n[!] Инициализация системы DedSec...")
	fakeHackAnimation()

	apps := []AppConfig{
		{
			Name:     "Chrome",
			Path:     "./apps/YChromeSetup.exe",
			Args:     []string{"/silent", "/install"},
			CheckCmd: `reg query "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\chrome.exe"`,
			CheckPath: "C://Program Files//Google//Chrome//Application//chrome.exe",
		},
		{
			Name:     "Telegram",
			Path:     "./apps/tsetup-x64.5.15.4.exe",
			Args:     []string{"/silent"},
			CheckCmd: `reg query "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\{53F49750-6209-4FBF-9CA8-7A333C87D1ED}"`,
			CheckPath: "C://Users//%USERNAME%//AppData//Roaming//Telegram Desktop//Telegram.exe",
		},
		{
			Name:     "Discord",
			Path:     "./apps/DiscordSetup.exe",
			Args:     []string{"/S"},
			CheckCmd: `reg query "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\Discord"`,
			CheckPath: "C://Users//%USERNAME%//AppData//Local//Discord//Update.exe",
		},
		{
			Name:     "Steam",
			Path:     "./apps/SteamSetup.exe",
			Args:     []string{"/S"},
			CheckCmd: `reg query "HKLM\SOFTWARE\WOW6432Node\Valve\Steam"`,
			CheckPath: "C://Program Files (x86)//Steam//steam.exe",
		},
		{
			Name:     "WinRAR",
			Path:     "./apps/WinRARSetup.exe",
			Args:     []string{"/S"},
			CheckCmd: `reg query "HKLM\SOFTWARE\WinRAR"`,
			CheckPath: "C://Program Files//WinRAR//WinRAR.exe",
		},
	}

	startTime := time.Now()
	var wg sync.WaitGroup

	for _, app := range apps {
		wg.Add(1)
		go func(a AppConfig) {
			defer wg.Done()
			installApp(a)
		}(app)
	}

	wg.Wait()
	totalDuration = time.Since(startTime)

	showInstallationSummary()
	pause()
}

func installApp(app AppConfig) {
	installMutex.Lock()
	defer installMutex.Unlock()

	if isInstalled(app) {
		fmt.Printf("\n✔ %s уже установлен (пропуск)\n", app.Name)
		successCount++
		return
	}

	if app.Path == "" {
		fmt.Printf("❌ [%s] Не указан путь к установщику\n", app.Name)
		failedCount++
		return
	}

	// Проверяем, существует ли файл установщика
	if _, err := os.Stat(app.Path); os.IsNotExist(err) {
		fmt.Printf("❌ [%s] Файл не найден: %s\n", app.Name, app.Path)
		failedCount++
		return
	}

	fmt.Printf("\n🔧 [%s] Начало установки...\n", time.Now().Format("15:04:05"))
	fmt.Printf(">> Исполняемый файл: %s\n", app.Path)

	startTime := time.Now()
	cmd := exec.Command(app.Path, app.Args...)

	// Запускаем процесс
	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ [%s] Ошибка запуска: %v\n", app.Name, err)
		failedCount++
		return
	}

	// Канал для отслеживания завершения
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Таймаут 10 минут на установку
	select {
	case err := <-done:
		duration := time.Since(startTime).Round(time.Second)
		if err != nil {
			fmt.Printf("❌ [%s] Ошибка установки (за %s): %v\n", app.Name, duration, err)
			failedCount++
		} else {
			fmt.Printf("✅ [%s] Успешно установлен за %s\n", app.Name, duration)
			successCount++
			installedApps[app.Name] = true
		}
	case <-time.After(10 * time.Minute):
		cmd.Process.Kill()
		fmt.Printf("⚠️ [%s] Превышено время ожидания (10 мин)\n", app.Name)
		failedCount++
	}
}

func isInstalled(app AppConfig) bool {
	if app.CheckCmd != "" {
		cmd := exec.Command("cmd", "/C", app.CheckCmd)
		if cmd.Run() == nil {
			return true
		}
	}
	if app.CheckPath != "" {
		if _, err := os.Stat(app.CheckPath); err == nil {
			return true
		}
	}
	return false
}

func showDedSecArt() {
	fmt.Print(`
  _____          _    _____ ______ 
 |  __ \   /\   | |  / ____|  ____|
 | |  | | /  \  | | | (___ | |__   
 | |  | |/ /\ \ | |  \___ \|  __|  
 | |__| / ____ \| |  ____) | |____ 
 |_____/_/    \_\_| |_____/|______|
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

	fmt.Println("\n\x1b[34m[+] Инициализация руткита DedSec_v9...\x1b[0m")
	time.Sleep(1 * time.Second)

	for phaseNum, phase := range phases {
		fmt.Printf("\n\x1b[36m[%d/%d] %s...\x1b[0m\n", phaseNum+1, len(phases), phase.name)

		for i := 0; i < 100; {
			step := rand.Intn(15) + 5
			if i+step > 100 {
				i = 100
			} else {
				i += step
			}

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

			target := targets[rand.Intn(len(targets))]
			ip := fmt.Sprintf("%d.%d.%d.%d:%d",
				rand.Intn(255), rand.Intn(255),
				rand.Intn(255), rand.Intn(255),
				rand.Intn(65535))

			fmt.Printf("\r>> %-25s [%-20s] \x1b[33m%3d%%\x1b[0m",
				truncate(target, 25),
				ip,
				i)

			time.Sleep(time.Duration(phase.delay+time.Duration(rand.Intn(200))) * time.Millisecond)
		}

		fmt.Printf("\r\x1b[32m[+] %s УСПЕШНО\x1b[0m%-30s\n", phase.name, "")
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("\n\x1b[5;32m[!] СИСТЕМА СКОМПРОМЕТИРОВАНА\x1b[0m")
	time.Sleep(1 * time.Second)
	fmt.Println("\x1b[32m[+] Подготовка к установке...\x1b[0m")
	time.Sleep(2 * time.Second)
}

func showInstallationSummary() {
	fmt.Printf("\n\x1b[36m=== ИТОГИ УСТАНОВКИ ===\x1b[0m\n")
	fmt.Printf("Успешно: \x1b[32m%d\x1b[0m\n", successCount)
	fmt.Printf("Неудачно: \x1b[31m%d\x1b[0m\n", failedCount)
	fmt.Printf("Общее время: \x1b[33m%s\x1b[0m\n", totalDuration.Round(time.Second))
	fmt.Println("\n\x1b[32m[+] DedSec завершил операцию\x1b[0m")
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

func clearConsole() {
	fmt.Print("\033[H\033[2J")
}

func setGreenText() {
	fmt.Print("\033[32m")
}

func pause() {
	fmt.Print("\nНажмите Enter для выхода...")
	fmt.Scanln()
}
