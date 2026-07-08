# **Reaktor-Architekten: Spielanleitung**

Ein asymmetrisches, kooperatives Aufbau- und Kollisionsspiel für 2 Spieler.

## **0\. Das Spielfeld im Detail**
| Column 1 (Player 1 Side) | Column 2 (Player 1 Side) | Column 3 (Player 1 Side) | Column 4 (Player 1 Side) | Column 5 (Player 1 Side) | Column 6 (Player 2 Side) | Column 7 (Player 2 Side) | Column 8 (Player 2 Side) | Column 9 (Player 2 Side) | 
| :---- | :---: | :---: | :---: | :---: | :---: | :---- | :---: | :---: |
| out-of-bounds | slot | slot | slot | slot (with wall to the right) | slot (wired to Industrie and Reaktorbedarf) | slot (wired to Industrie) | slot (wired to Industrie) | slot (wired to Wohnviertel) |
| Zünder | slot | slot | slot | Turbine (wired to Reaktorbedarf) | slot | slot | slot | slot (wired to wohnviertel) |
| out-of-bounds | slot | slot | slot | slot (with wall on the right) | slot (wired to Bahn and to Reaktorbedarf) | slot (wired to Bahn) | slot (wired to Bahn) | slot (wired to Wohnviertel)

## **1\. Setting & Konzept**

Ihr seid die Doppelspitze eines experimentellen sowjetischen Kraftwerks in den 1970er Jahren. Euer gemeinsames Ziel ist es, die städtische Industrie mit Energie zu versorgen und dabei nicht die Anlage in die Luft zu jagen. Das eigentliche Problem ist jedoch nicht die Physik, sondern das Ministerium: Die Obrigkeit diktiert absurde Produktionsziele (Quoten) und knappe Budgets, die selten zur Realität in eurem Reaktor passen.  
Das Spiel ist **asymmetrisch**:

* **Spieler 1 (Der Anlagenbauer):** Kontrolliert die linke Spielfeldhälfte (den **Reaktor**). Seine Aufgabe ist es, Thermodynamik und Kernphysik zu nutzen, um Wärme-Energie zu erzeugen und in die Mitte zu leiten.  
* **Spieler 2 (Der Netz-Experte):** Kontrolliert die rechte Spielfeldhälfte (das **Stromnetz**). Seine Aufgabe ist es, die eintreffende Energie aus der Mitte aufzufangen, umzuwandeln und sicher an die Stadt am rechten Spielfeldrand zu liefern.

## **2\. Spielaufbau**

* **Der Rahmen:** Das Spiel findet auf einem Hexagon-Raster statt (siehe "Das Spielfeld im Detail"). Links ist die Spielfeldhälfte für Spieler 1, rechts die Spielfeldhälfte für Spieler 2\. Um den Rand von der rechten Hälfte sind Zonen verteilt. Jede Zone hat Bedarfe und Schaden.
* **Zeit-Slider:** Aufgeteilt in Schicht und Monat. So viele Schichten und Monate habt ihr bereits gespielt  
* **Geldscheine**: Zum Merken wie viel Budget Spieler 1 und 2 diesen Monat noch zur Verfügung stehen  
* **Der Ursprung (Reaktor-Emitter):** Am äußersten linken Rand liegt der Zünder. Hier startet in jeder Schicht (Runde) die Kettenreaktion.  
* **Die Schnittstelle (Hauptturbine):** Genau in der Mitte des Rasters liegt die Turbine. Sie ist die Senke für den Reaktor und gleichzeitig der Startpunkt für das Stromnetz. Chips auf der Turbine zählen als ungebundene Spannungs-Chips des Spielers 2  
* **Die Bürokratie:** Legt die zwei gemischten Kartenstapel (**Ministerium für Finanzen** und **Ministerium für Energiewirtschaft**) bereit. Energiekarten sind in 3 Level eingeteilt. Sortiert euren Stapel so dass zuerst die zwei Level 1 Karten, dann zwei Level 2 Karten und dann 3 Level 3 Karten kommen.  
* **Richtungswürfel**: Wann immer euch die Regeln auffordern Chips in eine zufällige Richtung abzuschießen, würfelt. Die Zahl bestimmt die Kante. Tipp: Bei mehreren Chips, deren Richtung offen ist, könnt ihr gleich mehrere Würfel auf einmal würfeln (1 pro Chip), um schneller zu simulieren.

## **3\. Die Ressourcen ("State by Location" & Ladung)**

Das Spiel nutzt eine einzige Art von generischen Chips. Ein Chip erhält seine Bedeutung durch den Ort, an dem er liegt:

* **Als Ladung (Vorrat):** Liegen Chips fest auf einem Hexagon-Feld (z. B. auf einem Kohle-Feld), repräsentieren sie den unverbrauchten Brennstoff dieses Feldes.  
* **Im Flug (Trigger):** Fliegt ein Chip durch das Raster, ist er ein aktiver Auslöser (im Reaktor: **Wärme** oder **Neutron** / im Stromnetz: **Spannung**).

## **4\. Der Spielablauf (Die Planwirtschaft)**

Das Spiel verläuft in Monaten. Ein Monat besteht aus 5 Wochen-Schichten. Zu Beginn jedes Monats deckt ihr je eine Karte vom **Finanz-Stapel** und vom **Energie-Stapel**. Diese Kombination ist das absolute Gesetz für den gesamten Monat.  
**1\. Die Plan-Vorgabe:**

* Die offene **Energie-Karte** gibt den genauen Schichtplan vor, den das Städtische Netz in diesem Monat verlangt, sowie technische Sonderregeln. Schaut auf den Schichtplan, verteilt Bedarfschips auf die Spielfeldränder der Seite des Spielers 2\.  
* Die offene **Finanz-Karte** gibt an, wie viel **Geld** ihr beide zu Beginn *jeder einzelnen Schicht* dieses Monats erhaltet. (Gespartes Geld aus dem Monat davor verfällt). Nehmt euch die entsprechende Anzahl geldscheine für diesen Monat

**2\. Kaufen & Bauen (Start der Schicht):** Ihr kauft Felder aus dem Markt und legt sie in euer Raster. Schiebt die Geldschieber entsprechend der Kosten nach links. Sobald ein Feld gelegt wird, wird es mit Chips aus dem allgemeinen Vorrat als "Ladung" befüllt. Spieler 2 kann auch 1 Geld pro Schadenschip ausgeben um Schaden von Zonen zu entfernen.
**3\. Zünden (Kostenlos):** Spieler 1 feuert am Zünder (am linken Rand) genau **1 Basis-Trigger** (Wärme-Chip oder Neutron) in eine gewählte Richtung ab. Spieler 2 feuert an der Turbine oder befriedigt direkt einen Bedarf im Kraftwerk  
**4\. Simulation:** Trifft ein fliegender Chip auf ein Feld, passiert Folgendes:

* *Flugbahn:*   
  * Chips fliegen in geraden Linien  
  * Spieler 1: Prallt ein Wärme-Chip gegen die feste Außenwand von Spieler 1 prallt er im gleichen Winkel ab. Andere Chips (Neutronen) verpuffen.  
  * Spieler 2: Jeder Spannungs-Chip, der einen Rand trifft, verschwindet sofort und vernichtet dort einen Bedarfs-Chip. Ist kein Bedarfschip vorhanden, wird der Chip stattdessen auf das Schadensbereich der Zone gelegt. Chips auf den Schadensbereich zählen auch zum Limit von 8 für Spieler 2. Ihr müsst überschüssigen Strom zwingend durch Speicher-Felder, Erdungs-Felder oder andere Ränder vernichten, sonst droht durch die Reflektion sofort die Kritische Masse (\>8 Chips).  
* *Reaktion:*   
  * *Treffer auf reguläres Feld*: Das Feld löst seinen Effekt aus (z.B. wird zerstört oder 1 Wärme wird zu 2 Wärme). Um das Lösen von Energie darzustellen, wird die richtige Anzahl Chips aus dem Bereich “gebunden” des Feldes in den Bereich “ungebunden” gelegt. Der eingehende Chip landet auch im Bereich “ungebunden”.  
  * *Treffer auf ausgebranntes Feld:* Der Chip verschwindet  
  * *Treffer auf Zünder*: Der Chip wird vernichtet. 
  * *Treffer auf Turbine:*   
    * Wärme-Chip: Wird zu einem ungebundenen Spannungs-Chip. Spieler 2 wählt wann und in welche richtung er die Spannungs-Chips abschießt. Sie zählen aber zum 8 Chip Limit und können deshalb nicht endlos aufgestaut werden.  
    * Spannungs-Chip: Zählt wie der Kraftwerk-Spielfeldrand. Verschwindet also und entfernt einen Bedarf.  
* *Kritische Masse:* Existieren zu irgendeinem Zeitpunkt mehr als **8 *gelöste* Chips gleichzeitig** auf der linken Spielhälfte (Spieler 1\) *oder* mehr als 8 *gelöste* Chips auf der rechten Seite habt ihr beide verloren. Im Flug befindliche Chips zählen als gelöst.  
* *Beliebige Reihenfolge:* Jeder Spieler darf in seiner Spielhälfte frei entscheiden, in welcher Reihenfolge Aktionen ausgespielt werden.  
* *Quoten-Erfüllung & Überlastung:* Die Bedarfs-Chips die nicht befriedigt wurden, bleiben am Ende einer Schicht einfach liegen. In der nächsten Schicht kommen die neunen Bedarfschips vom Schichtplan einfach dazu. 

**6\. Schichtabschluss**: Wurde während der Simulation der letzte Ladungs-Chip von einem Feld vom Bereich “gebunden” entfernt, wird das Feld jetzt abgeräumt.  
**7\. Monatsabschluss:** Ist der Monat vorbei und es gibt keine Bedarfschips mehr auf den Rändern der Spielfeldhälfte des Spielers zwei, habt ihr den Monat überlebt. Legt das überschüssige Geld zurück. Deckt die nächste Finanz- und Energiekarten auf. Sind die Stapel leer (nach 6 erfolgreichen Monaten (einem Halbjahr)) gewinnt ihr das Spiel. Befinden sich noch Bedarfe auf dem Spielfeld habt ihr verloren.

## **5\. Ministeriums-Karten**

### **Ministerium für Energiewirtschaft (Die Quote)**

---

#### **Eröffnungsfeier \- Stufe 1**

**Sonderregel:** Nur schow: Generatoren kostet 1 Chip weniger  
**Schichtplan:**

| Schicht | Industrie (Oben) | Wohnviertel (Rechts) | Bahn (Unten) | Kraftwerk (Links / Eigenbedarf) |
| :---- | :---: | :---: | :---: | :---: |
| **Schicht 1** | 1 | 1 | 0 | **1** |
| **Schicht 2** | 2 | 1 | 0 | **1** |
| **Schicht 3** | 1 | 2 | 0 | **2** |
| **Schicht 4** | 3 | 2 | 1 | **2** |
| **Schicht 5** | 2 | 2 | 1 | **2** |

*Kontext: Ein hochrangiger Parteifunktionär besucht die Anlage; die Sicherheitsprotokolle werden vorübergehend gelockert, um den guten Schein zu wahren.*

#### **Optimierung des lokalen Netzes \- Stufe 1**

**Sonderregel:** Transformatoren kosten 1 Geld statt 2 Geld  
**Schichtplan:**

| Schicht | Industrie (Oben) | Wohnviertel (Rechts) | Bahn (Unten) | Kraftwerk (Links / Eigenbedarf) |
| :---- | :---: | :---: | :---: | :---: |
| **Schicht 1** | 1 | 1 | 1 | **1** |
| **Schicht 2** | 2 | 1 | 0 | **1** |
| **Schicht 3** | 1 | 2 | 0 | **1** |
| **Schicht 4** | 2 | 1 | 1 | **1** |
| **Schicht 5** | 1 | 2 | 1 | **2** |

*Historischer Kontext: Um die Effizienz zu steigern, wurden in den 70ern oft lokale Netzverbesserungen vorgenommen, um die Verluste bei der Fernübertragung zu reduzieren. Diese Maßnahmen waren oft erfolgreich, schufen aber Abhängigkeiten bei der Leitungsführung.*

#### **"Die technologische Transformation" \- Stufe 2**

**Sonderregel:** Uran-Platten kosten diesen Monat 4 Geld statt 5

**Schichtplan:**

| Schicht | Industrie (Oben) | Wohnviertel (Rechts) | Bahn (Unten) | Kraftwerk (Links / Eigenbedarf) |
| :---- | :---: | :---: | :---: | :---: |
| **Schicht 1** | 2 | 2 | 0 | **0** |
| **Schicht 2** | 3 | 1 | 0 | **1** |
| **Schicht 3** | 2 | 2 | 1 | **1** |
| **Schicht 4** | 3 | 1 | 0 | **1** |
| **Schicht 5** | 2 | 1 | 1 | **1** |

*Historischer Kontext: Das Ministerium drängt auf die Nutzung neuer, unerprobter Brennelemente, um die Kapazität zu erhöhen. Uran wird subventioniert*

#### **"Der Gossnab-Liefervertrag" \- Stufe 2**

**Sonderregel:** Minderwertige Lieferungen. Alle neu gebauten Kohle-Brennkammern und Transformatoren starten in diesem Jahr mit 1 Ladung weniger.  
**Schichtplan:**

| Schicht | Industrie (Oben) | Wohnviertel (Rechts) | Bahn (Unten) | Kraftwerk (Links / Eigenbedarf) |
| :---- | :---: | :---: | :---: | :---: |
| **Schicht 1** | 1 | 1 | 0 | **0** |
| **Schicht 2** | 2 | 0 | 1 | **1** |
| **Schicht 3** | 0 | 3 | 1 | **1** |
| **Schicht 4** | 1 | 1 | 0 | **0** |
| **Schicht 5** | 2 | 1 | 1 | **1** |

*Historischer Kontext: Das Versorgungskomitee (Gossnab) maß seinen Erfolg oft in Tonnen. Kraftwerke erhielten daher regelmäßig Millionen Tonnen minderwertiger oder klatschnasser Kohle, da diese schwerer war und die Transportquote der Bahn erfüllte – auch wenn sie kaum brannte.*

#### **"Testlauf unter Volllast" \- Stufe 3**

**Sonderregel:** Die Obrigkeit duldet keine Verzögerung. Alle *Absorber-Stäbe* und *Kühltürme* auf dem Spielfeld verlieren für dieses Jahr ihre Funktion.

**Schichtplan:**

| Schicht | Industrie (Oben) | Wohnviertel (Rechts) | Bahn (Unten) | Kraftwerk (Links / Eigenbedarf) |
| :---- | :---: | :---: | :---: | :---: |
| **Schicht 1** | 2 | 2 | 1 | **1** |
| **Schicht 2** | 3 | 1 | 1 | **1** |
| **Schicht 3** | 2 | 2 | 1 | **1** |
| **Schicht 4** | 3 | 1 | 1 | **1** |
| **Schicht 5** | 2 | 1 | 2 | **1** |

*Historischer Kontext: Die Katastrophe von Tschernobyl (1986) ereignete sich während eines staatlich verordneten Sicherheitstests. Da die Netzbetreiber in Kiew unerwartet Strom forderten, wurde der Test verzögert und schließlich unter massivem Druck mit deaktivierten Sicherheitssystemen durchgeführt.*

#### 

#### **"Schturmowschtschina (Die Sturmarbeit)" \- Stufe 3**

**Sonderregel:** Wir hinken dem Plan hinterher\! Spieler 1 *muss* am Anfang jeder Schicht 2 Wärme-Chips (statt 1\) in den Reaktor feuern.  
**Schichtplan:**

| Schicht | Industrie (Oben) | Wohnviertel (Rechts) | Bahn (Unten) | Kraftwerk (Links / Eigenbedarf) |
| :---- | :---: | :---: | :---: | :---: |
| **Schicht 1** | 1 | 1 | 1 | **1** |
| **Schicht 2** | 2 | 2 | 1 | **1** |
| **Schicht 3** | 4 | 2 | 2 | **2** |
| **Schicht 4** | 6 | 1 | 2 | **2** |
| **Schicht 5** | 7 | 2 | 1 | **2** |

*Historischer Kontext: Da Material in der UdSSR oft erst spät im Monat geliefert wurde, standen Fabriken wochenlang still. In der letzten Woche brach die "Schturmowschtschina" aus – ein wahnsinniger, ungesicherter Produktionsrausch, um die Monatsquote noch irgendwie zu erfüllen.*

### **Um jeden Preis**

Die "Kritische Masse" liegt in diesem Monat bei 10 statt 8 Chips (auf beiden Seiten). 

### **Ministerium für Finanzen (Das Budget)**

**1\. "Triumph der Schwerindustrie"**

* **Schicht-Budget:** Reaktor: 3 Geld | Stromnetz: 3 Geld.  
* **Sonderregel:** Uran ist um 1 Geld günstiger.

**2\. "Nationale Sparmaßnahmen"**

* **Schicht-Budget:** Reaktor: 1 Geld | Stromnetz: 1 Geld.  
* **Sonderregel:** Keine. Ihr müsst mit dem Schrott arbeiten, den ihr habt.

**3\. Nukleares Wettrüsten**

* **Schicht-Budget:** Reaktor: 2 Geld | Stromnetz: 3 Geld.  
* **Sonderregel:** Reparaturen werden nicht bewilligt. Leere (ausgebrannte) Felder dürfen in diesem Jahr nicht mit neuen Feldern überbaut werden.  
* *Historischer Kontext: In der Anlage Majak fiel 1957 das Kühlsystem für nukleare Abfälle aus. Da das Finanzbüro keine Reparaturmittel bewilligte (um die Produktion nicht zu stören), überhitzte ein Tank und löste den drittschwersten Nuklearunfall der Geschichte aus (Kyshtym-Vorfall).*

## **6\. Sektor 1: Der Reaktor (Felder für Spieler 1\)**

* **Feld entfernen (Kosten: 1):** Entferne ein beliebiges Feld  
* **Ablenk-Spiegel (Kosten: 1 Geld | Ladung: Keine):** Lenkt eintreffende Teilchen im fixen Winkel ab.  
* **Kohle-Brennkammer (Kosten: 2 Geld | Ladung: 4 Chips):** 1 Wärme trifft ein \-\> verbraucht 1 Ladung \-\> feuert 2 Wärme zufällig ab. Vernichtet einkommende Chips im ausgebrannten Zustand.  
* **Kühlturm (Kosten: 2 Geld | Ladung: Keine):** Vernichtet eintreffende Wärme restlos (Notbremse).  
* **Erdgas-Kessel (Kosten: 3 Geld | Ladung: 8 Chips | ab 2\. Monat):** 1 Wärme trifft ein \-\> verbraucht 3 Ladung \-\> feuert 4 Wärme ab. Vernichtet einkommende Chips im ausgebrannten Zustand.  
* **Absorber-Stab (Kosten: 3 Geld | Ladung: Keine):** Vernichtet eintreffende Neutronen restlos.  
* **Uran-Platte (Kosten: 5 Geld | Ladung: 2 Chips | ab 3\. Monat):** 1 *Neutron* trifft ein \-\> verbraucht 1 Ladung \-\> feuert 2 Neutronen & 1 Wärme ab. Trifft 1 Wärme ein so wird sie in eine zufällige richtung weitergeschickt.
* **Tokamak-Kammer (Kosten: 8 Geld | Ladung: Unendlich | ab 4\. Monat):** 4 Neutronen treffen ein \-\> feuert 8 Wärme ab

## **7\. Sektor 2: Das Stromnetz (Felder für Spieler 2\)**

* **Feld entfernen (Kosten: 1): Entferne ein beliebiges Feld**  
* **Relais / Weiche (Kosten: 1 Geld | Ladung: Keine):** Lenkt eintreffende Spannung im fixen Winkel ab.  
* **Transformator (Kosten: 2 Geld | Ladung: 4 Chips):** 1 Spannung trifft ein (hochspannung) \-\> verbraucht 1 Ladung \-\> feuert 2 Spannung in zufällige Richtung ab (niederspannung). Ein ausgebrannter Trafo vernichtet jeden einkommende Chips  
* **Erdung / Widerstand (Kosten: 2 Geld | Ladung: 4 Chips):** Leitet eintreffende Spannung ab (vernichtet sie). Verbraucht 1 Ladung. Wichtig bei Überproduktion\!  
* **Notgenerator (Kosten: 3 Geld | Ladung 2 Chips):** Schiese jederzeit, wenn Spieler 2 es wünscht, die Ladung in eine gewünschte Richtung. Trifft ein Spannungs-Chip ein, wird der Generator sofort zerstört. Entferne ihn inklusive aller seiner ladungen sofort.  
* **Kondensator-Bank (Kosten: 4 Geld | Ladung: Maximal 5 Chips | ab 2\. Monat):** Nimmt bis zu 5 Spannungs-Chips auf. Diese können einzeln *innerhalb einer Schicht* in eine Richtung der Wahl geschossen werden. Leert sich beim Schichtwechsel komplett. Empfängt der Kondensator mehr als 5 Spannung explodiert er. Nimm das Feld vom Spielgrid und schieße alle Chips in zufällige Richtungen.  Chips im Kondensator zählen nicht zum Limit von 8 Chips für die linke Spielhälfte.
* **Blei-Akkumulator (Kosten: 3 Geld | Ladung: Maximal 3 Chip | ab 2\. Monat):** Wie Kondensator-Bank, aber verliert zu Begin der Schicht 1 Chip anstatt geleert zu werden. Bewege einen Chip auf ungebunden. Dieser Chip fliegt unkontrolliert in eine zufällige Richtung. Statt zu explodieren, werden einkommende Ladungschips in eine zufällige Richtung umgeleitet, falls der speicher voll ist (Spannungs-Spike)  
* **Pumpspeicherwerk (Kosten: 4 Geld | Ladung: Maximal 5 Chips | ab 3\. Monat):** Wie Kondensator-Bank, aber muss beim Schichtwechsel überhaupt nicht geleert werden. Statt zu explodieren, wird Spannung in eine zufällige Richtung umgeleitet, falls der speicher voll ist (Spannungs-Spike)  
* **Hochspannungs-Kaskade (Kosten: 3 Geld | Ladung: 8 Chips | Ab 3\. Monat):** 1 Spannung (hochspannung) trifft ein \-\> verbraucht 3 Ladung \-\> feuert 4 Spannung in zufällige Richtungen ab (niederspannung).  
* **Supraleiter (Kosten: 4 Geld | ab 4\. Monat):** Zeigt auf einen beliebigen Spielfeldrand. Jede eintreffende Spannung erreicht sofort diesen Rand

