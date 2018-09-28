package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type resource struct {
	jsonres string
	jsonmax string
	label   string
	unit    string
}

var tab = map[string]resource{
	"Name":  {"v.name", "", "Name", ""},
	"Throt": {"f.throttle", "", "Throttle", ""},
	"H":     {"n.heading", "", "Heading", "°"},
	"P":     {"n.pitch", "", "Pitch", "°"},
	"R":     {"n.roll", "", "Roll", "°"},
	"ToPro": {"v.angleToPrograde", "", "ToPrograde", "°"},
	"Vel":   {"v.surfaceVelocity", "", "Surface velocity", "m/s"},
	"OVel":  {"v.orbitalVelocity", "", "Orbital velocity", "m/s"},
	"G":     {"v.geeForce", "", "Gee", "G"},
	"Atm":   {"v.atmosphericDensity", "", "Atmos density", ""},
	"Q":     {"v.dynamicPressure", "", "Q", ""},
	"Alt":   {"v.altitude", "", "Radar altitude", ""},
	"Pe":    {"o.PeA", "", "Pe", ""},
	"Ap":    {"o.ApA", "", "Ap", ""},
	"TTPe":  {"o.timeToPe", "", "Time to Pe", ""},
	"TTAp":  {"o.timeToAp", "", "Time to Ap", ""},
	"Incl":  {"o.inclination", "", "Inclination", "°"},
	"Ecc":   {"o.eccentricity", "", "Eccentricity", ""},
	"St":    {"mj.node", "", "Stage", ""},

	"SAS":  {"v.sasValue", "", "SAS", ""},
	"RCS":  {"v.rcsValue", "", "RCS", ""},
	"LGT":  {"v.lightValue", "", "LIGHT", ""},
	"BRK":  {"v.brakeValue", "", "BRK", ""},
	"GEAR": {"v.gearValue", "", "GEAR", ""},

	"Kero":  {"r.resource[Kerosene]", "r.resourceMax[Kerosene]", "Kerosene", "L"},
	"LOX":   {"r.resource[LqdOxygen]", "r.resourceMax[LqdOxygen]", "Liquid oxygen", "L"},
	"Hydra": {"r.resource[Hydrazine]", "r.resourceMax[Hydrazine]", "Hydrazine", "L"},
	"Aero":  {"r.resource[Aerozine50]", "r.resourceMax[Aerozine50]", "Aerozine 50", "L"},
	"NTO":   {"r.resource[NTO]", "r.resourceMax[NTO]", "NTO", "L"},
	"MMH":   {"r.resource[MMH]", "r.resourceMax[MMH]", "MMH", "L"},
	"Xen":   {"r.resource[XenonGas]", "r.resourceMax[XenonGas]", "Xenon gas", "L"},
	"UDMH":  {"r.resource[UDMH]", "r.resourceMax[UDMH]", "UDMH", "L"},
	"Mono":  {"r.resource[MonoPropellant]", "r.resourceMax[MonoPropellant]", "Monopropellant", "L"},
	"Elec":  {"r.resource[ElectricCharge]", "r.resourceMax[ElectricCharge]", "Electric charge", "Wh"},
	"Solid": {"r.resource[SolidFuel]", "r.resourceMax[SolidFuel]", "Solid fuel", "kg"},

	//"Warp": {"t.timeWarp", "", "Warp", ""},
	"T": {"v.missionTime", "", "Time", ""},
}

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func printtime(vals map[string]interface{}, k string) {
	pre := "T+"
	var v float64
	var ok bool
	if v, ok = vals[k].(float64); !ok {
		return
	}
	if v < 0 {
		pre = "T-"
		v = -v
	}
	s := math.Mod(v, 60)
	m := math.Mod(v/60, 60)
	h := math.Mod(v/60/60, 60)
	d := math.Mod(v/60/60/24, 24)
	fmt.Printf("%20s  %s %dd %dh %02dm %02ds\n", tab[k].label, pre, int(d), int(h), int(m), int(s))
}

func printval(vals map[string]interface{}, k string) {
	if _, ok := vals[k]; !ok {
		return
	}
	fmt.Printf("%20s %10.2f %s\n", tab[k].label, vals[k], tab[k].unit)
}

func printdist(vals map[string]interface{}, k string) {
	var v float64
	var ok bool

	if v, ok = vals[k].(float64); !ok {
		return
	}
	u := "m"
	if v > 1000000 || v < -1000000 {
		v /= 1000000
		u = "Mm"
	} else if v > 1000 || v < -1000 {
		v /= 1000
		u = "km"
	}
	fmt.Printf("%20s %10.2f %s\n", tab[k].label, v, u)
}

func drawbar(v, max float64) {
	ful := int(math.Ceil(40 * (v / max)))
	emp := int(math.Floor(40 * (1 - v/max)))
	for i := 0; i < ful; i++ {
		fmt.Printf("█")
	}
	for i := 0; i < emp; i++ {
		fmt.Printf("-")
	}
}

func printresource(vals map[string]interface{}, k string) {
	if _, ok := vals[k].(float64); !ok {
		return
	}
	if vals[k].(float64) == -1 {
		return
	}
	max := 1.0
	if v, ok := vals[k+"max"].(float64); ok {
		max = v
	}
	fmt.Printf("%20s %10.2f %3s  %6.2f%%  ", tab[k].label, vals[k].(float64), tab[k].unit, 100*(vals[k].(float64)/max))

	drawbar(vals[k].(float64), max)
	fmt.Println()
}

func printpercent(vals map[string]interface{}, k string) {
	if _, ok := vals[k]; !ok {
		return
	}
	fmt.Printf("%20s %d%%\n", tab[k].label, int(vals[k].(float64)*100))
}

func printbool(vals map[string]interface{}, k string) {
	var b, ok bool
	if b, ok = vals[k].(bool); !ok {
		return
	}
	if b {
		fmt.Printf("[%s]", tab[k].label)
	} else {
		fmt.Printf(" %s ", tab[k].label)
	}
}

func printorient(vals map[string]interface{}) {
	var ok bool
	h, ok := vals["H"].(float64)
	p, ok := vals["P"].(float64)
	r, ok := vals["R"].(float64)
	_, ok = vals["ToPro"].(float64)
	if !ok {
		return
	}
	fmt.Printf("%20s % 07.2f° % 07.2f° % 07.2f°\n", "h p r", h, p, r)
	fmt.Printf("%20s % 07.2f°\n", "Ang to prograde", vals["ToPro"])
}

func refresh() {
	url := "http://localhost:80/telemachus/datalink?"
	i := 0
	for k, r := range tab {
		url += fmt.Sprintf("%s=%s", k, r.jsonres)
		if r.jsonmax != "" {
			url += fmt.Sprintf("&%smax=%s", k, r.jsonmax)
		}
		if i < len(tab)-1 {
			url += "&"
		}
		i++
	}

	resp, err := http.Get(url)
	if err != nil {
		clear()
		fmt.Println("no signal")
		return
	}

	txt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		clear()
		fmt.Println(err)
		return
	}

	var vals map[string]interface{}

	err = json.Unmarshal(txt, &vals)
	if err != nil {
		clear()
		fmt.Println(err)
		return
	}

	clear()

	printtime(vals, "T")
	fmt.Println()
	printval(vals, "St")
	fmt.Printf("                    ")
	printbool(vals, "SAS")
	printbool(vals, "RCS")
	printbool(vals, "LGT")
	printbool(vals, "BRK")
	printbool(vals, "GEAR")
	fmt.Println()
	printpercent(vals, "Throt")
	printorient(vals)
	printval(vals, "Vel")
	printval(vals, "OVel")
	printval(vals, "G")
	printdist(vals, "Alt")
	printval(vals, "Atm")
	printval(vals, "Q")

	fmt.Println()
	printdist(vals, "Ap")
	printdist(vals, "Pe")
	printtime(vals, "TTAp")
	printtime(vals, "TTPe")
	printval(vals, "Incl")
	printval(vals, "Ecc")

	fmt.Println()

	printresource(vals, "Elec")
	printresource(vals, "Kero")
	printresource(vals, "LOX")
	printresource(vals, "Hydra")
	printresource(vals, "Aero")
	printresource(vals, "NTO")
	printresource(vals, "MMH")
	printresource(vals, "UDMH")
	printresource(vals, "Xen")
	printresource(vals, "Mono")
	printresource(vals, "Solid")
}

func main() {
	for {
		refresh()
		time.Sleep(50 * time.Millisecond)
	}
}
