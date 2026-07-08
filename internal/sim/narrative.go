package sim

import (
	"fmt"
	"strings"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
)

func dirName(travelDir int) string {
	return hex.DisplayDirName(travelDir)
}

func narrate(event string, resolved *Chip, b *board.State, queue []Chip, emitted []Chip) string {
	switch event {
	case "Start":
		return narrateStart(queue)
	case "verloren":
		return "Kritische Masse ueberschritten - das Spiel ist verloren."
	case "Schrittlimit":
		return fmt.Sprintf("Simulation nach %d Schritten abgebrochen (vermutlich Endlosschleife).", MaxStepsPerRun)
	case "Ende":
		return "Schichtende: Warteschlange leer, keine freiwilligen Schüsse mehr moeglich."
	case "Waerme-Reflektion":
		if resolved != nil {
			return fmt.Sprintf(
				"Wärme prallt an der Außenwand ab und fliegt in Richtung %s.",
				dirName(resolved.Dir),
			)
		}
		return narrateWithoutChip(event, b, queue)
	case "Spannungs-Spike":
		if resolved != nil {
			return fmt.Sprintf("Spannungs-Spike: Spannung wird in Richtung %s reflektiert.", dirName(resolved.Dir))
		}
		return narrateWithoutChip(event, b, queue)
	case "Zuender-Treffer":
		if resolved != nil {
			return fmt.Sprintf("%s trifft den Zünder und wird vernichtet.", chipName(resolved.Type))
		}
		return "Ein Chip trifft den Zünder und wird vernichtet."
	case "Freiwilliger Schuss":
		return narrateWithoutChip(event, b, queue)
	}
	if resolved != nil {
		if text := narrateBorderDemand(event, *resolved); text != "" {
			return text
		}
		if text := narrateBorderDamage(event, *resolved); text != "" {
			return text
		}
	}
	if resolved == nil {
		return narrateWithoutChip(event, b, queue)
	}
	var parts []string
	if resolved.Pos.IsEmitter() {
		parts = append(parts, fmt.Sprintf(
			"Spieler 1 schießt %s vom Zünder in Richtung %s.",
			chipAccusative(resolved.Type), dirName(resolved.Dir),
		))
	}
	if detail := narrateResolved(event, *resolved, b, queue, emitted); detail != "" {
		parts = append(parts, detail)
	}
	if len(parts) == 0 {
		return event
	}
	return strings.Join(parts, " ")
}

func narrateStart(queue []Chip) string {
	emitter := hex.Coord{Q: hex.EmitterCol, R: hex.EmitterRow}
	var chips []Chip
	for _, c := range queue {
		if c.Pos == emitter {
			chips = append(chips, c)
		}
	}
	if len(chips) == 0 {
		return "Schichtbeginn: keine Chips am Zünder."
	}
	parts := make([]string, 0, len(chips))
	for _, c := range chips {
		parts = append(parts, fmt.Sprintf("%s in Richtung %s", chipName(c.Type), dirName(c.Dir)))
	}
	return fmt.Sprintf("Schichtbeginn: am Zünder bereit — %s.", strings.Join(parts, ", "))
}

func narrateWithoutChip(event string, b *board.State, queue []Chip) string {
	switch event {
	case "Freiwilliger Schuss":
		for _, c := range queue {
			if c.Type != ChipVoltage {
				continue
			}
			t := b.Tiles[c.Pos.Q][c.Pos.R]
			if isVoluntarySource(t.Type) {
				return fmt.Sprintf(
					"Spieler 2 feuert Spannung vom %s in Richtung %s.",
					fieldVon(t.Type), dirName(c.Dir),
				)
			}
		}
		return "Spieler 2 feuert Spannung freiwillig ab."
	case "Waerme-Reflektion":
		for _, c := range queue {
			if c.Type == ChipHeat {
				return fmt.Sprintf("Wärme prallt an der Außenwand ab und fliegt in Richtung %s.", dirName(c.Dir))
			}
		}
		return "Wärme prallt an der Außenwand ab."
	case "Spannungs-Spike":
		for _, c := range queue {
			if c.Type == ChipVoltage {
				return fmt.Sprintf("Spannungs-Spike: Spannung wird in Richtung %s reflektiert.", dirName(c.Dir))
			}
		}
		return "Spannungs-Spike an der Außenwand."
	}
	return event
}

func narrateBorderDamage(event string, chip Chip) string {
	z, ok := board.ZoneFromBorderDamageEvent(event)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s erzeugt Schaden in %s (kein Bedarf frei).", chipName(chip.Type), z.String())
}

func narrateBorderDemand(event string, chip Chip) string {
	z, ok := board.ZoneFromBorderDemandEvent(event)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s erfüllt Rand-Bedarf %s.", chipName(chip.Type), z.String())
}

func narrateResolved(event string, chip Chip, b *board.State, queue []Chip, emitted []Chip) string {
	target := chip.Pos.Neighbor(chip.Dir)
	if !target.Valid() {
		return ""
	}
	tile := b.Tiles[target.Q][target.R]

	switch event {
	case "Feldreaktion", "Notgenerator zerstoert":
		return narrateFieldHit(chip, target, tile, queue, emitted, event == "Notgenerator zerstoert")
	case "Turbine":
		emitted := emittedAt(queue, target)
		if chip.Type == ChipHeat {
			if len(emitted) > 0 {
				return fmt.Sprintf("Wärme trifft die Turbine. Die Turbine erzeugt Spannung in Richtung %s.", dirName(emitted[0].Dir))
			}
			return "Wärme trifft die Turbine."
		}
		if chip.Type == ChipVoltage {
			return "Spannung trifft die Turbine."
		}
	case "Leeres Feld":
		return fmt.Sprintf("%s landet auf leerem Feld und wartet in Richtung %s.", chipName(chip.Type), dirName(chip.Dir))
	case "Bedarf erfuellt":
		zone := demandZoneLabel(b, target)
		if zone != "" {
			return fmt.Sprintf("%s erfüllt Bedarf %s.", chipName(chip.Type), zone)
		}
		return fmt.Sprintf("%s erfüllt einen Bedarfs-Chip.", chipName(chip.Type))
	case "Ausgebrannt":
		return fmt.Sprintf("%s trifft ausgebranntes Feld (%s) und verpufft.", chipName(chip.Type), fieldShortName(tile.Type))
	case "Innere Wand":
		return fmt.Sprintf("%s wird von der inneren Wand gestoppt.", chipName(chip.Type))
	case "Waerme verpufft":
		return "Wärme verpufft an der Außenwand von Spieler 2."
	case "Neutron verpufft":
		return "Neutron verpufft an der Außenwand."
	case "Spannung verpufft":
		return "Spannung verpufft außerhalb des Stromnetzes."
	}
	return ""
}

func narrateFieldHit(chip Chip, target hex.Coord, tile field.Tile, queue []Chip, emitted []Chip, destroyed bool) string {
	name := fieldShortName(tile.Type)
	incoming := chipName(chip.Type)
	released := emitted
	if len(released) == 0 {
		released = emittedAt(queue, target)
	}

	if destroyed {
		return fmt.Sprintf("%s trifft den Notgenerator — er wird sofort zerstört.", incoming)
	}

	switch tile.Type {
	case field.CoalChamber, field.GasBoiler, field.Transformer, field.HVCascade:
		if len(released) > 0 && len(emitted) > 0 {
			return fmt.Sprintf(
				"%s trifft %s. %s schießt %s ab. %s",
				incoming, name, capitalize(name), formatChipCount(released), formatDirRolls(released),
			)
		}
		if len(released) > 0 && len(emitted) == 0 {
			return fmt.Sprintf("%s trifft %s, reagiert aber nicht.", incoming, name)
		}
		return fmt.Sprintf("%s trifft %s, reagiert aber nicht.", incoming, name)

	case field.UraniumPlate:
		if chip.Type == ChipHeat && len(released) == 1 {
			return fmt.Sprintf(
				"%s trifft Uran und wird abgelenkt. %s",
				incoming, formatDirRolls(released),
			)
		}
		if chip.Type == ChipNeutron && len(released) > 0 {
			neutrons := filterType(released, ChipNeutron)
			heats := filterType(released, ChipHeat)
			var parts []string
			if len(neutrons) > 0 {
				parts = append(parts, fmt.Sprintf("%d Neutronen (%s)", len(neutrons), dirList(neutrons)))
			}
			if len(heats) > 0 {
				parts = append(parts, fmt.Sprintf("%d Wärme (%s)", len(heats), dirList(heats)))
			}
			return fmt.Sprintf(
				"%s trifft Uran. Uran feuert %s ab.",
				incoming, strings.Join(parts, " und "),
			)
		}
		return fmt.Sprintf("%s trifft Uran.", incoming)

	case field.Tokamak:
		if chip.Type == ChipNeutron {
			if len(released) > 0 {
				return fmt.Sprintf(
					"Tokamak zündet nach 4 Neutronen und schießt %s ab. %s",
					formatChipCount(released), formatDirRolls(released),
				)
			}
			return fmt.Sprintf("Neutron trifft Tokamak (%d/4 bis zur Zündung).", tile.TokamakCounter)
		}

	case field.Mirror:
		if chip.Type == ChipVoltage {
			return fmt.Sprintf("%s trifft %s und passiert ihn.", incoming, name)
		}
		return narrateWireField(incoming, "den Spiegel", chip, wireEmittedChip(chip, tile))

	case field.Relay:
		if chip.Type != ChipVoltage {
			return fmt.Sprintf("%s trifft %s und passiert ihn.", incoming, name)
		}
		return narrateWireField(incoming, "das Relais", chip, wireEmittedChip(chip, tile))

	case field.CoolingTower:
		if chip.Type == ChipHeat {
			return "Wärme trifft Kühlturm und wird vernichtet."
		}

	case field.AbsorberRod:
		if chip.Type == ChipNeutron {
			return "Neutron trifft Absorber-Stab und wird vernichtet."
		}

	case field.Ground:
		if chip.Type == ChipVoltage && len(released) == 0 {
			return fmt.Sprintf("Spannung trifft Erdung und wird abgeleitet (Ladung %d).", tile.Charge)
		}

	case field.CapacitorBank, field.PumpedStorage, field.LeadAccumulator:
		if chip.Type == ChipVoltage && len(released) == 0 {
			return fmt.Sprintf("Spannung wird im %s gespeichert (%d Ladung).", name, tile.StoredVoltage)
		}
		if len(released) > 0 {
			return fmt.Sprintf("%s ist voll — Spannungs-Spike in Richtung %s.", name, dirName(released[0].Dir))
		}

	case field.Superconductor:
		if len(released) == 1 {
			return fmt.Sprintf("Spannung teleportiert über Supraleiter und fliegt in Richtung %s weiter.", dirName(released[0].Dir))
		}
	}

	if len(released) > 0 {
		return fmt.Sprintf(
			"%s trifft %s. %s",
			incoming, name, formatDirRolls(released),
		)
	}
	if tile.BurnedOut {
		return fmt.Sprintf("%s trifft %s und verpufft.", incoming, name)
	}
	return fmt.Sprintf("%s trifft %s.", incoming, name)
}

func emittedAt(queue []Chip, pos hex.Coord) []Chip {
	out := make([]Chip, 0)
	for _, c := range queue {
		if c.Pos == pos {
			out = append(out, c)
		}
	}
	return out
}

func filterType(chips []Chip, t ChipType) []Chip {
	out := make([]Chip, 0)
	for _, c := range chips {
		if c.Type == t {
			out = append(out, c)
		}
	}
	return out
}

func formatChipCount(chips []Chip) string {
	if len(chips) == 0 {
		return "nichts"
	}
	counts := map[ChipType]int{}
	for _, c := range chips {
		counts[c.Type]++
	}
	parts := make([]string, 0, 3)
	for _, t := range []ChipType{ChipHeat, ChipNeutron, ChipVoltage} {
		if n := counts[t]; n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, chipName(t)))
		}
	}
	return strings.Join(parts, " und ")
}

func formatDirRolls(chips []Chip) string {
	if len(chips) == 0 {
		return ""
	}
	dirs := make([]string, len(chips))
	for i, c := range chips {
		dirs[i] = dirName(c.Dir)
	}
	if len(dirs) == 1 {
		return fmt.Sprintf("Richtungs-Wurf ergibt %s.", dirs[0])
	}
	return fmt.Sprintf("Richtungs-Würfe ergeben %s.", strings.Join(dirs, " und "))
}

func dirList(chips []Chip) string {
	parts := make([]string, len(chips))
	for i, c := range chips {
		parts[i] = dirName(c.Dir)
	}
	return strings.Join(parts, ", ")
}

func chipName(t ChipType) string {
	switch t {
	case ChipNeutron:
		return "Neutron"
	case ChipVoltage:
		return "Spannung"
	default:
		return "Wärme"
	}
}

func chipAccusative(t ChipType) string {
	switch t {
	case ChipNeutron:
		return "ein Neutron"
	case ChipVoltage:
		return "eine Spannung"
	default:
		return "eine Wärme"
	}
}

func fieldShortName(t field.Type) string {
	switch t {
	case field.CoalChamber:
		return "die Kohle"
	case field.GasBoiler:
		return "das Erdgas"
	case field.UraniumPlate:
		return "Uran"
	case field.Tokamak:
		return "den Tokamak"
	case field.CoolingTower:
		return "den Kühlturm"
	case field.AbsorberRod:
		return "den Absorber-Stab"
	case field.Mirror:
		return "den Spiegel"
	case field.Relay:
		return "das Relais"
	case field.Transformer:
		return "den Transformator"
	case field.Ground:
		return "die Erdung"
	case field.EmergencyGenerator:
		return "den Notgenerator"
	case field.CapacitorBank:
		return "die Kondensator-Bank"
	case field.PumpedStorage:
		return "das Pumpspeicherwerk"
	case field.LeadAccumulator:
		return "den Blei-Akkumulator"
	case field.HVCascade:
		return "die Hochspannungs-Kaskade"
	case field.Superconductor:
		return "den Supraleiter"
	default:
		if info, ok := field.Catalog[t]; ok {
			return info.Name
		}
		return "ein Feld"
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] -= 'a' - 'A'
	}
	return string(r)
}

func demandZoneLabel(b *board.State, c hex.Coord) string {
	label := b.DemandLabel(c)
	if label == "" {
		return ""
	}
	switch label[0] {
	case 'I':
		return "Industrie"
	case 'W':
		return "Wohnviertel"
	case 'b':
		return "Bahn"
	case 'R':
		return "Reaktoreigenbedarf"
	default:
		return label
	}
}

func narrateWireField(incoming, fieldName string, chip Chip, emitted Chip) string {
	if emitted.Dir == chip.Dir {
		return fmt.Sprintf("%s fliegt durch %s in Richtung %s.", incoming, fieldName, dirName(emitted.Dir))
	}
	return fmt.Sprintf("%s trifft %s und wird in Richtung %s gelenkt.", incoming, fieldName, dirName(emitted.Dir))
}

func wireEmittedChip(chip Chip, tile field.Tile) Chip {
	incoming := (chip.Dir + 3) % 6
	return Chip{Type: chip.Type, Dir: tile.Orientation.WireOutgoing(incoming)}
}

func isVoluntarySource(t field.Type) bool {
	switch t {
	case field.CapacitorBank, field.PumpedStorage, field.LeadAccumulator, field.EmergencyGenerator:
		return true
	default:
		return false
	}
}

func fieldVon(t field.Type) string {
	switch t {
	case field.EmergencyGenerator:
		return "Notgenerator"
	case field.CapacitorBank:
		return "Kondensator-Bank"
	case field.PumpedStorage:
		return "Pumpspeicherwerk"
	case field.LeadAccumulator:
		return "Blei-Akkumulator"
	default:
		if info, ok := field.Catalog[t]; ok {
			return info.Name
		}
		return "Speicher"
	}
}
