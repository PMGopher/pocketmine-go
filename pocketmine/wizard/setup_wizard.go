// Package wizard is a port of pocketmine\wizard: the set-up wizard run on the first start.
package wizard

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/server"
	"pocketmine-go/pocketmine/utils"
)

// SetupWizard is a port of pocketmine\wizard\SetupWizard.
type SetupWizard struct {
	dataPath string
	lang     *lang.Language

	in  *bufio.Reader
	out io.Writer
	// sleep is PHP's sleep() (tests replace it).
	sleep func(time.Duration)
	// getIP and getInternalIP are Internet::getIP/getInternalIP (tests replace them).
	getIP         func() (string, bool)
	getInternalIP func() (string, error)
}

// NewSetupWizard is a port of SetupWizard::__construct: it reads from stdin and writes to stdout.
func NewSetupWizard(dataPath string) *SetupWizard {
	return NewSetupWizardWithIO(dataPath, os.Stdin, os.Stdout)
}

// NewSetupWizardWithIO is NewSetupWizard with the given input and output.
func NewSetupWizardWithIO(dataPath string, in io.Reader, out io.Writer) *SetupWizard {
	return &SetupWizard{
		dataPath:      dataPath,
		in:            bufio.NewReader(in),
		out:           out,
		sleep:         time.Sleep,
		getIP:         func() (string, bool) { return utils.GetIP(false) },
		getInternalIP: utils.GetInternalIP,
	}
}

// Run is a port of SetupWizard::run: false if the server shouldn't start (the license wasn't
// accepted).
func (w *SetupWizard) Run() bool {
	w.message(pocketmine.Name + " set-up wizard")

	langs, err := lang.GetLanguageList("")
	if err != nil {
		w.error("No language files found, please use provided builds or clone the repository recursively.")
		return false
	}

	w.message("Please select a language")
	codes := make([]string, 0, len(langs))
	for short := range langs {
		codes = append(codes, short)
	}
	sort.Strings(codes)
	for _, short := range codes {
		w.writeLine(" " + langs[short] + " => " + short)
	}

	var selected string
	for selected == "" {
		selected = strings.ToLower(w.getInput("Language", "eng", ""))
		if _, ok := langs[selected]; !ok {
			w.error("Couldn't find the language")
			selected = ""
		}
	}

	if w.lang, err = lang.NewLanguage(selected, "", ""); err != nil {
		w.error(err.Error())
		return false
	}

	w.message(w.lang.Translate(lang.KnownTranslationFactory.LanguageHasBeenSelected()))

	if !w.showLicense() {
		return false
	}

	//This has to happen here to prevent user avoiding agreeing to license
	config, err := utils.NewConfig(filepath.Join(w.dataPath, "server.properties"), utils.ConfigProperties, nil)
	if err != nil {
		w.error(err.Error())
		return false
	}
	config.Set(server.PropertyLanguage, selected)
	_ = config.Save()

	if strings.ToLower(w.getInput(w.lang.Translate(lang.KnownTranslationFactory.SkipInstaller()), "n", "y/N")) == "y" {
		w.printIpDetails()
		return true
	}

	w.writeLine("")
	w.welcome()

	w.generateBaseConfig(config)
	w.generateUserFiles(config)
	w.networkFunctions(config)
	_ = config.Save()

	w.printIpDetails()

	w.endWizard()

	return true
}

// showLicense is a port of SetupWizard::showLicense.
func (w *SetupWizard) showLicense() bool {
	w.message(w.lang.Translate(lang.KnownTranslationFactory.WelcomeToPocketmine(pocketmine.Name)))
	fmt.Fprint(w.out, `
  This program is free software: you can redistribute it and/or modify
  it under the terms of the GNU Lesser General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

`)
	w.writeLine("")
	if strings.ToLower(w.getInput(w.lang.Translate(lang.KnownTranslationFactory.AcceptLicense()), "n", "y/N")) != "y" {
		w.error(w.lang.Translate(lang.KnownTranslationFactory.YouHaveToAcceptTheLicense(pocketmine.Name)))
		w.sleep(5 * time.Second)

		return false
	}

	return true
}

// welcome is a port of SetupWizard::welcome.
func (w *SetupWizard) welcome() {
	w.message(w.lang.Translate(lang.KnownTranslationFactory.SettingUpServerNow()))
	w.message(w.lang.Translate(lang.KnownTranslationFactory.DefaultValuesInfo()))
	w.message(w.lang.Translate(lang.KnownTranslationFactory.ServerProperties()))
}

// askPort is a port of SetupWizard::askPort.
func (w *SetupWizard) askPort(prompt *lang.Translatable, def int) int {
	for {
		port := phpInt(w.getInput(w.lang.Translate(prompt), strconv.Itoa(def), ""))
		if port <= 0 || port > 65535 {
			w.error(w.lang.Translate(lang.KnownTranslationFactory.InvalidPort()))
			continue
		}
		return port
	}
}

// generateBaseConfig is a port of SetupWizard::generateBaseConfig.
func (w *SetupWizard) generateBaseConfig(config *utils.Config) {
	config.Set(server.PropertyMotd, w.getInput(w.lang.Translate(lang.KnownTranslationFactory.NameYourServer()), server.DefaultServerName, ""))

	w.message(w.lang.Translate(lang.KnownTranslationFactory.PortWarning()))

	config.Set(server.PropertyServerPortIPv4, w.askPort(lang.KnownTranslationFactory.ServerPortV4(), server.DefaultPortIPv4))
	config.Set(server.PropertyServerPortIPv6, w.askPort(lang.KnownTranslationFactory.ServerPortV6(), server.DefaultPortIPv6))

	w.message(w.lang.Translate(lang.KnownTranslationFactory.GamemodeInfo()))

	gamemode := ""
	for gamemode == "" {
		switch phpInt(w.getInput(w.lang.Translate(lang.KnownTranslationFactory.DefaultGamemode()), "0", "")) {
		case 0:
			gamemode = "SURVIVAL"
		case 1:
			gamemode = "CREATIVE"
		}
	}
	//TODO: this probably shouldn't use the enum name directly
	config.Set(server.PropertyGameMode, gamemode)

	config.Set(server.PropertyMaxPlayers, phpInt(w.getInput(w.lang.Translate(lang.KnownTranslationFactory.MaxPlayers()), strconv.Itoa(server.DefaultMaxPlayers), "")))

	config.Set(server.PropertyViewDistance, phpInt(w.getInput(w.lang.Translate(lang.KnownTranslationFactory.ViewDistance()), strconv.Itoa(server.DefaultMaxViewDistance), "")))
}

// generateUserFiles is a port of SetupWizard::generateUserFiles.
func (w *SetupWizard) generateUserFiles(config *utils.Config) {
	w.message(w.lang.Translate(lang.KnownTranslationFactory.OpInfo()))

	op := strings.ToLower(w.getInput(w.lang.Translate(lang.KnownTranslationFactory.OpWho()), "", ""))
	if op == "" {
		w.error(w.lang.Translate(lang.KnownTranslationFactory.OpWarning()))
	} else if ops, err := utils.NewConfig(filepath.Join(w.dataPath, "ops.txt"), utils.ConfigEnum, nil); err == nil {
		ops.Set(op, true)
		_ = ops.Save()
	}

	w.message(w.lang.Translate(lang.KnownTranslationFactory.WhitelistInfo()))

	if strings.ToLower(w.getInput(w.lang.Translate(lang.KnownTranslationFactory.WhitelistEnable()), "n", "y/N")) == "y" {
		w.error(w.lang.Translate(lang.KnownTranslationFactory.WhitelistWarning()))
		config.Set(server.PropertyWhitelist, true)
	} else {
		config.Set(server.PropertyWhitelist, false)
	}
}

// networkFunctions is a port of SetupWizard::networkFunctions.
func (w *SetupWizard) networkFunctions(config *utils.Config) {
	w.error(w.lang.Translate(lang.KnownTranslationFactory.QueryWarning1()))
	w.error(w.lang.Translate(lang.KnownTranslationFactory.QueryWarning2()))
	if strings.ToLower(w.getInput(w.lang.Translate(lang.KnownTranslationFactory.QueryDisable()), "n", "y/N")) == "y" {
		config.Set(server.PropertyEnableQuery, false)
	} else {
		config.Set(server.PropertyEnableQuery, true)
	}
}

// printIpDetails is a port of SetupWizard::printIpDetails.
func (w *SetupWizard) printIpDetails() {
	w.message(w.lang.Translate(lang.KnownTranslationFactory.IpGet()))

	externalIP, ok := w.getIP()
	if !ok {
		externalIP = "unknown (server offline)"
	}
	internalIP, err := w.getInternalIP()
	if err != nil {
		internalIP = "unknown (" + err.Error() + ")"
	}

	w.error(w.lang.Translate(lang.KnownTranslationFactory.IpWarning(externalIP, internalIP)))
	w.error(w.lang.Translate(lang.KnownTranslationFactory.IpConfirm()))
	w.readLine()
}

// endWizard is a port of SetupWizard::endWizard.
func (w *SetupWizard) endWizard() {
	w.message(w.lang.Translate(lang.KnownTranslationFactory.YouHaveFinished()))
	w.message(w.lang.Translate(lang.KnownTranslationFactory.PocketminePlugins()))
	w.message(w.lang.Translate(lang.KnownTranslationFactory.PocketmineWillStart(pocketmine.Name)))

	w.writeLine("")
	w.writeLine("")

	w.sleep(4 * time.Second)
}

func (w *SetupWizard) writeLine(line string) {
	eol := "\n"
	if os.PathSeparator == '\\' {
		eol = "\r\n" // PHP_EOL
	}
	fmt.Fprint(w.out, line+eol)
}

func (w *SetupWizard) readLine() string {
	line, _ := w.in.ReadString('\n')
	return strings.TrimSpace(line)
}

func (w *SetupWizard) message(message string) { w.writeLine("[*] " + message) }

func (w *SetupWizard) error(message string) { w.writeLine("[!] " + message) }

// getInput is a port of SetupWizard::getInput.
func (w *SetupWizard) getInput(message, def, options string) string {
	message = "[?] " + message

	if options != "" || def != "" {
		shown := options
		if shown == "" {
			shown = def
		}
		message += " (" + shown + ")"
	}
	message += ": "

	fmt.Fprint(w.out, message)

	input := w.readLine()
	if input == "" {
		return def
	}
	return input
}

// phpInt is PHP's (int) cast of a string: the leading integer, or 0.
func phpInt(value string) int {
	value = strings.TrimSpace(value)
	end := 0
	for end < len(value) && (value[end] >= '0' && value[end] <= '9' || end == 0 && (value[end] == '-' || value[end] == '+')) {
		end++
	}
	i, _ := strconv.Atoi(value[:end])
	return i
}
