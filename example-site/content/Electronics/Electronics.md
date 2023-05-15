# Electronics
* Wikipedia [Electronics](https://en.wikipedia.org/wiki/Electronics)
* Book: [Make: Electronics](https://www.amazon.com/Make-Electronics-Charles-Platt-ebook/dp/B0B3LS5K2Z/) by Charles Platt (MELEC)

# Multimeters
* Wikipedia [Multimeter](https://en.wikipedia.org/wiki/Multimeter)
* Features to look for:
  * Transistor tester: symbol hFE and has circle of slots with openings for transistor leads.
  * Continuity testing: diode symbol.
  * AC amp measurement.
  * Resistance: ideally up to 20 MΩ
  * Measures capacitance: Although, less common and more expensive.
* MELEC author favorite is BK Precision Test Bench 388B (p 5), although it currently (2022) lists for $160 US.

# Voltage
* Wikipedia [Voltage](https://en.wikipedia.org/wiki/Voltage)

# Current
* Wikipedia [Electric current](https://en.wikipedia.org/wiki/Electric_current)
* Circuits are drawn with current going from positive to negative, even though electrons move in the opposite direction.
  This is called "__conventional current__", and is drawn this way for historical reasons, before it was known that it's the
  electrons that move. For calculations it makes no difference. A negative charge moving in one direction is equivalent to
  a positive charge moving in the other direction.

# Power
* Wikipedia [Electric power](https://en.wikipedia.org/wiki/Electric_power)
* P = VI

# Resistance
* Wikipedia [Resistor](https://en.wikipedia.org/wiki/Resistor)
* Wikipedia [Electronic color code: Resistors](https://en.wikipedia.org/wiki/Electronic_color_code#Resistors)
  <img class="illustration" src="{{STATIC}}/electronics/resistor-colors.png" />
* Pure water has high resistance. Impurities such as salt give it conductance.
* Typical resistors are made to handle up to ¼ watt.

# Photoresistor <span id=photoresistor />
* Wikipedia [Photoresistor](https://en.wikipedia.org/wiki/Photoresistor):
  "(also known as a Photocell, or light-dependent resistor, __LDR__, or
  photo-conductive cell) is a passive component that decreases resistance with
  respect to receiving luminosity (light) on the component's sensitive surface.
  __The resistance of a photoresistor decreases with increase in incident light
  intensity__"
* Range varies from LDR to LDR. Measure for each.
* Place in series with another resistor to create a voltage divider circuit. 
* Compute optimal resistance of the 2nd resistor with this equation, where
  LDR<sub>bright</sub> is resistance at max light and LDR<sub>dark</sub> is
  resistance with no light (AVR-PROG p 131):
  <img class="illustration" src="{{STATIC}}/electronics/ldr-formula.png" />

# Capacitors
* Wikipedia [Farad](https://en.wikipedia.org/wiki/Farad)
* Wikipedia [Capacitance](https://en.wikipedia.org/wiki/Capacitance)
* Wikipedia [Capacitor](https://en.wikipedia.org/wiki/Capacitor)
* Wikipedia [Capacitor: Capacitor markings](https://en.wikipedia.org/wiki/Capacitor#Capacitor_markings):
  "Smaller capacitors, such as ceramic types, often use a shorthand-notation
  consisting of three digits and an optional letter, where the digits (XYZ)
  denote the capacitance in picofarad (pF), calculated as XY × 10Z, and the
  letter indicating the tolerance. Common tolerances are ±5%, ±10%, and ±20%,
  denotes as J, K, and M, respectively."
* Wikipedia [RC circuit](https://en.wikipedia.org/wiki/RC_circuit)
* A resistor and capacitor in series is called an RC circuit. (MELEC p 73)
* The term mF is not used much, probably to not be confused with μF. (MELEC p 74)
* Ceramic capacitors do not have __polarity__. Electrolytic capacitors do; the
  longer leg is the positive terminal, or look for stripe with - on it showing
  negative terminal.
* Datasheet key specs:
  * Working voltage: Max voltage that can be applied. Capacitor could blow or explode at higher voltages.
    * 5v is not uncommon for __electrolytic capacitors__. 
    * Make a note of this when ordering __ceramic capacitors__ as its often not listed 
      anywhere. A typical value is 25v, but it can be lower. (MELEC p 78)
* __Be very careful with larger electrolytic capacitors__. They can store lethal amounts of charge for a long time; e.g. inside antique TVs.
* Capacitors block current in a DC circuit. However, an initial current does pass when the 
  capacitor is initially charged. This can be significant when a capacitor is charged quickly.
  __[Capacitive coupling](https://en.wikipedia.org/wiki/Capacitive_coupling)__ is the name of this 
  effect. (MELEC p 83)

# LEDs
* Flat side is negative (the cathode).
* Or, longer lead is positive (the anode).
* Datasheet key values:
  * "Absolute maximum ratings":
    * __I<sub>F</sub>__: __Forward current__; e.g. 50mA. Never exceed this.
  * "Typical characteristics":
    * __Forward Voltage__: Is the suggested value that the LED operate at.
      Usually given as a range; e.g. 2.1v-26v. Will be listed with a suggested
      value for I<sub>F</sub>; e.g.  2.1v for 20mA. Keep current at this value.

# Switches
* Wikipedia [Switch](https://en.wikipedia.org/wiki/Switch)
* __Pole__: independent circuit.
* __Throw__: connection terminal.
* Common types:
  * Single pole, single throw (SPST or 1P1T): Basic on/off switch.
  * Single pole, double throw (SPDT or 1P2T): Two on positions, one off.
  * Double pole, double throw (DPDT or 2P2T): Two separate circuits (poles) each controlled with the same physical switch.
    Each circuit has two on terminals and on off terminal.
* __Momentary switches__: are spring-loaded so they snap back to their default position when pressure is released.
* __Normally open__ (NO): Contacts are normally open.
* __Normally closed__ (NC): Contacts are normally closed.
* Inductive loads (e.g. a motor) can cause a __spark__ when switched off. Make sure
  switch is rated at twice regular amperage to handle this, or switch will
  eventually burn out. (MELEC p 57)

# Relays
* Wikipedia [Relay](https://en.wikipedia.org/wiki/Relay)
* Usually a relay has a __default position when power is not applied__. 
* __Latching__ relays remain in the position they've been switched to when turned off. These are less common.
* Schematics show non-latching relays in their off position.
* Key __datasheet__ specs:
  * __Switching capacity__: Max current the relay can switch. This is for passive loads, for inductive loads
    half this. (See "spark" in switches above.) A "small-signal relay" can't switch much current. (MELEC p 64)
  * __Coil voltage__: ideal voltage to apply to energize the relay.
  * __Set voltage__: the min voltage to close the switch.
  * __Operating current__: the power consumption of the coil when the switch is set.
* __The coil only has a mechanical connection with the contacts__, and no
  electrical connection.  This allows the power supply for the coil to have no
  connection to the power supply of the device you are switching on and off.
  (MELEC p 196)
* Use a [freewheeling diode](https://en.wikipedia.org/wiki/Flyback_diode) to prevent reverse current surges
  when the coil is turned off (MELEC p 202). The diode blocks normal forward current but allows the reverse
  current to flow back through it, instead of to components in the rest of the circuit that could be damaged (e.g. transistors).
  <img class="illustration" src="{{STATIC}}/electronics/freewheeling-diode.png" />

# Transistors
* Wikipedia [Transistor](https://en.wikipedia.org/wiki/Transistor)
* Structure:
  * Wikipedia [MOSFET](https://en.wikipedia.org/wiki/MOSFET)
    * Is more efficient than bi-polar transistors. (MELEC p 91)
  * Wikipedia [Bipolar junction transistor](https://en.wikipedia.org/wiki/Bipolar_junction_transistor)
    * Are either __NPN__ or __PNP__.
    * Are less vulnerable to accidents than MOSFETs. (MELEC p 91)
    * Was invented before the MOSFET. (MELEC p 91)
* In schematics __arrow shows direction of conventional current__.
* __NPN transistors__ are drawn with arrow leaving emitter. Increasing base voltage ("P") promotes current flow.
  Current flows from collector to emitter.
  <img class="illustration" src="{{STATIC}}/electronics/npn-schematic.png" />
* __PNP transistors__ are drawn with arrow entering collector. Decreasing base voltage ("N") promotes current flow.
  Current flows from emitter to collector.
  img class="illustration" src="{{STATIC}}/electronics/pnp-schematic.png" />
* A simple way to remember how arrow is drawn: The arrow in an NPN is "Never Pointing iN."
* A transistor __amplifies current__. The current through the transistors is an amplified value
  of the current through the base. The amplification factor is called the __beta__. (MELEC p 90)
* Beta can be measured with a multimeter.
* Beta tends to be linear, which is important for applications such as music amplification.
* Transistors compared to relays (MELEC p 92):
  <img class="illustration" src="{{STATIC}}/electronics/transistors-and-relays.png" />

# Transistors: MOSFET
* [MOSFET](https://en.wikipedia.org/wiki/MOSFET)
  * Pins:
    * Source: Connect to ground.
    * Drain: Connect to circuit that will be switched on and off.
    * Gate: Connect to input signal. When signal is high the switch is closed, and when it's low it's open.
* Wikipedia [2N7000](https://en.wikipedia.org/wiki/2N7000): 
  * Is a MOSFET.
  * Is for low-power switching applications.
  * The gate draws almost no current and so doesn't need a resistor.
  * Looking at it with flat side towards you and pins down, the far left pin is the source (connect to ground), the far right is the drain, and 
    the middle is the gate.

# Transistors: 2N3904
* Used lots in MELEC.
* "It is widely used in small low-voltage circuits, is rated for up to 40V, and is able to pass up to 200mA." (MELEC p 87)
* Wikipedia [2N3904](https://en.wikipedia.org/wiki/2N3904): 
  "is a common __NPN__ bipolar junction transistor used for general-purpose
  low-power amplifying or switching applications.[1][2][3] It is designed for
  low current and power, medium voltage, and can operate at moderately high
  speeds."
* With flat side facing you, collector (current flows in) is on the right and
  emitter (current flows out) is on the left. (MELEC p 87)
* Voltage on base just needs to be higher than voltage on emitter for current to flow. (MELEC p 88)
* The base must be 0.7v higher than the emitter for current to flow. (MELEC p 90)
* There's always a voltage drop going across the transistor, from collector to emitter. This
  is called __effective resistance__. The effective resistance lowers as the voltage of base goes up. 
  When the effective resistance has gone as low as it can the transistor is set to be __saturated__. (MELEC p 90)
* Always limit current flowing through a transistor, to protect it, similar to what's done for an LED.

# Transistors: Common emitter versus common collector
* A transistor and the load it drives can be arranged so that the load is connected
  to either the emitter or collector.
  <img class="illustration" src="{{STATIC}}/electronics/common-emitter-collector.png" />
* Wikipedia [Common emitter](https://en.wikipedia.org/wiki/Common_emitter)
* Wikipedia [Common collector](https://en.wikipedia.org/wiki/Common_collector)
* The common collector arrangement is more intuitive since the transistor is allowing
  power to the load. However, the common emitter configuration actually supplies more
  voltage to the load. (MELEC p 106-109)

# Breadboards
* Wire colors used (MELEC p 71):
  * Red: For connections to power supply (voltage is voltage of power supply).
  * Blue: Ground.
  * Black: In-between.

# Soldering
* Ideal soldering iron:
  * 30w melts solder faster and better but can damage components. 15w can be good for learning.
  * He likes the Weller Therma-Boost (MELEC p 110).
* The solder must be listed as for electronics. Craft and plumbing solder can have acid that damages electronics.
* Use a crumpled up paper to clean the tip, rubbing it quickly against the paper and applying a bit of solder
  until it's smooth. You want to make sure there's no carbon on the tip. (MELEC p 119)
* Apply manual force after the solder cools to test.

# Power Supplies
* Buy a 9v adapter and clip the end off to solder 22 gauge solid wire to each wire. 
  Make the negative lead shorter than the positive before soldering, to reduce the chance
  of a short if the two wires were ever to touch. (MELEC p 125-127)

# Perfboard
* Wikipedia [Perfboard](https://en.wikipedia.org/wiki/Perfboard)
* youtube.com [Circuit Skills: Perfboard Prototyping](https://www.youtube.com/watch?v=3N3ApzmyjzE)
* Perfboard without copper is great to start. Or, another good option is a perfboard with 
  copper connections that mirror the connections found on a breadboard; see figure 7-15 on MELEC p 168.
* 6" x 8" is a good size. Use a hack saw to size the board to your circuit. (A wood saw
  will be dulled by the fiberglass.)
* Diagram where components will be placed on the front. Take a picture and then
  using image editing software flip the diagram to create another diagram
  showing how wires will be run on the back. See MELEC fig 14-2 on p 131 for an example.
* Perfboard holes are 0.1" apart. ICs called "through-hole" have pins that are
  0.1" apart. The packaging for this type of chip is called __dual-inline__
  packaging. __Surface mount__ chip pins, on the other hand, are closer together.
  Part numbers will often have an S in them if they're surface mount.  Avoid
  chips with part numbers that start with S, SMT, SMD, SO, SOIC, SSOP, TSSOP,
  or TVSOP.
* Example part number for a chip in the 7400 family. Note the N suffix for "dual-inline":
  <img class="illustration" src="{{STATIC}}/electronics/chip-part-numbers.png" />

# Project Boxes
* Epoxy glue can be used to mount LEDs and speakers. (MELEC p 172)
* Switches and push buttons usually come with a threaded neck and nut to fit.
* Use a caliper to measure exact distances, comparing parts to the project box.
* To drill holes in thin plastic (MELEC p 172):
  * Forstner drill bits create clean holes in thin plastic.
  * Another option is to use progressively large drill bits to make a hole the correct size.
  * Use a reamer drill attachment to make holes slightly larger, or a deburring tool. 
  * A countersink bit is useful too.
* Mounting circuit board to box (MELEC p 172):
  * Use #4 size machine bolts with washers and nylon-insert lock-nuts. 
  * Use nylon washers or other spacers to create room for wires under the perfboard.
* Aluminum boxes are fine, but require an insulating layer between the perfboard and box.
  A mouse pad cut to size is one option. (MELEC p 173)
* Multiple boards can by layered with bubble wrap. (MELEC p 173)

# Voltage Regulators
* [78xx](https://en.wikipedia.org/wiki/78xx)
  "is a family of self-contained fixed linear voltage regulator integrated
  circuits. The 78xx family is commonly used in electronic circuits requiring
  a regulated power supply due to their ease-of-use and low cost."
* __LM7805__ takes 7.5VDC to 12VDC and brings it down to 5VDC. (MELEC p 138)
* Use with __smoothing capacitors__ when output goes to logic chips (MELEC p 186):
  <img class="illustration" src="{{STATIC}}/electronics/lm7805.png" />

# 555 Timer <span id=555 />
* Wikipedia [555 timer IC](https://en.wikipedia.org/wiki/555_timer_IC):
  "is an integrated circuit (chip) used in a variety of timer, delay, pulse
  generation, and oscillator applications. Derivatives provide two (556) or
  four (558) timing circuits in one package."
* Input voltage can be between 5VDC and 16VDC. Output can be up to 200mA.
* The traditional 555 is
  [TTL](https://en.wikipedia.org/wiki/Transistor%E2%80%93transistor_logic)
  ("transistor-transistor logic", built from bipolar junction transistors).
  The advantage of this is it's cheap, robust, and can deliver up to 200 mA.
  However, the output has an initial spike. A smoothing capacitor can be added
  to the output to compensate. (MELEC p 150)
* There are [CMOS](https://en.wikipedia.org/wiki/CMOS) versions as well, but they're
  more susceptible to static and can't deliver as much power.
* Microcontrollers are often used now in place of a 555 timer, but they can't provide
  as much power and have to be programmed. (MELEC p 146)
* Pinout: <img class="illustration" src="{{STATIC}}/electronics/555-pinout.png" />
* Pins:
  * Trigger (pin 2, input): Keep high by default with a pull-up resistor. Ground to trigger.
  * Output (pin 3, output): Low by default.
  * Reset (pin 4, input): Keep high by default with a pull-up resistor. Ground to
    cause chip to reset. While low, chip does nothing, and output (pin 3) is low.
  * Control (pin 5):
  * Threshold (pin 6, input): 
  * Discharge (pin 7): Discharges capacitor attached to threshold (pin 6) into chip. 

# 555 Timer: __Monostable Mode__
* __Output pulse lasts a configured amount of time, and ends.__ (MELEC p 141)
* Pulse can last anywhere from around 10 milliseconds to 18 minutes. (MELEC p 144)
* Example (MELEC 143): <img class="illustration" src="{{STATIC}}/electronics/555-monostable.png" />
* Button A press causes trigger (pin 2) to go low, causing output (pin 3), and LED, to go high
  for a configurable amount of time. Or, if trigger is left low (button is held down), the output stays high.
* Potentiometer P1 and capacitor C1 determine the length of the pulse. C1 charges through P1 and R1,
  until it reaches a threshold, measured by Threshold (pin 6), and then it discharges into chip through Discharge (pin 7).
* See table 15-15 on p 144 (MELEC) for what to set P1+R1 and C1 to for a particular output pulse length.
* Reset (pin 4) is normally kept high. When it goes low, with button B press, any output signal (pin 3)
  goes low, and the chip does nothing until reset goes high again.
* Control (pin 5) has no purpose in this mode. The capacitor to ground just helps prevent voltage fluctuations.
* Inside 555 in monostable mode (MELEC p 145):
  <img class="illustration" src="{{STATIC}}/electronics/555-monostable-inside.png" />
* When Trigger (pin 2) goes low, comparator A pulls the flip-flop to its side, causing
  the Output (pin 3) to go high, and allowing the timing capacitor to charge. When the 
  charge gets high enough comparator B pulls the flip-flop to its side, bringing the
  Output low, and allowing the capacitor to discharge through Discharge (pin 7).
* On powerup the Output can sometimes emit a startup pulse. To workaround this, place a 
  1 μF capacitor between Reset (pin 4) and ground (fig 15-17 p 146). This will cause
  reset to be momentarily low on startup and prevent the output pulse.

# 555 Timer: Bi-stable Mode
* __Flip-flop mode.__ Output goes high with trigger and low with reset.(MELEC p 146)
* This mode can be used to debounce a noisy switch. (MELEC p 216)
* This is configured just like monostable but with all the timing components removed
  and Threshold (pin 6, input) connected to ground.

# 555 Timer: Astable Mode
* __Output oscillates.__

# 555 Timer: Schmitt Trigger Mode <span id=555-schmitt />
* __Output (digital) goes high and low depending on input (analog), but the input
  has to change enough to trigger a reversal.__  This property is called [hysteresis](https://en.wikipedia.org/wiki/Hysteresis).
* Wikipedia [Schmitt trigger](https://en.wikipedia.org/wiki/Schmitt_trigger)
  * "The circuit is named a trigger because the output retains its value until the input changes sufficiently to trigger a change."
  * "The Schmitt trigger is effectively a one bit analog to digital converter."
* Wikipedia __[555 timer: Schmitt trigger](https://en.wikipedia.org/wiki/555_timer_IC#Schmitt_trigger)__:
  Converts a noisy analog input into a clean digital output. 
* Example 1: A thermostat where the temperature can fluctuate around the set
  temperature, but you don't want a heater, for example, being turned
  rapidly on and off.)
* Example 2: A button push generates contact bounce the can be cleaned up to
  create a single digital output. (MELEC p 216)
* Example 3: Photocell circuit to detect light. The input is analog and based on the
  resistance provided by a Light Dependent Resistor. The output is binary, and goes high 
  lots of light but then doesn't change back to low until the input has changed enough,
  to avoid flicker at the set-point.
* Example of a [Photo Switch Circuit](https://www.circuitstoday.com/photo-switch-circuit):
  <img class="illustration" src="{{STATIC}}/electronics/555-schmitt.png" />
* Pin 2 is Trigger and its comparator pulls the internal flip-flop its way to connect Output (pin 3) to VCC when it goes low (less than 1/3 the input voltage.)
* Pin 6 is Threshold and its comparator pulls the internal flip-flop its way to connects Output (pin 3) to ground when it goes high (above 2/3 the input voltage.)
* As it gets dark the resistance of LDR goes up and so the voltage on pin 6 goes up, causing Output (pin 3) to go low.
* As it gets light the resistance of LDR goes down. The voltage on pin 2 goes down, causing Output to go high.
* The hysteresis happens because there's a gap between the voltages required to trigger the comparator,
  because of the internal 3 resistor voltage divider circuit.
* POT R2 determines the point around which the comparators are triggered, to vary how light or dark
  it is when they're triggered.

# Diodes
* Wikipedia [Diode](https://en.wikipedia.org/wiki/Diode)
* __Signal diodes__ are used for forward currents as high as 200mA. __Rectifier diodes__ can
  handle much more current. (MELEC p 154)
* He uses the [1N4148 signal diode](https://en.wikipedia.org/wiki/1N4148_signal_diode). (MELEC p 154)
* The __anode__ is the positive terminal and __cathode__ the negative.

# Seven-segment Display
* Wikipedia [Seven-segment display](https://en.wikipedia.org/wiki/Seven-segment_display):
  "is a form of electronic display device for displaying decimal numerals that
  is an alternative to the more complex dot matrix displays."

# 4026B Counter
* Wikipedia [4000-series integrated circuits](https://en.wikipedia.org/wiki/4000-series_integrated_circuits)
* "4026 – __Decade counter with 7-segment digit decoded output__."
* The 4026B is a CMOS chip and so can be easily damaged by static electricity. (MELEC p 175)
* Anti-static wrist straps can be good to prevent this, if you notice problems
  with electric shocks when you touch things. He's never had a problem with
  this, though. The alligator clip on the strap would ideally be connected to a steel or copper water pipe,
  but something such as a filing cabinet works too.
* Storage of CMOS chips in the foam packaging they came in is good because the foam is designed
  to distribute charge across all pins. Or, you can use aluminum foil.
* The 4026B has output pins that correspond the the pins of a seven-segment display. (MELEC p 179)
* The current from the pins should not exceed 1mA which he just goes with and things work but ideally 
  the pins would be fed through a darlington array of transistors to amplify the current and make
  the display bright. (MELEC p 179)
* One chip counts to 9. Additional chips can be daisy changed for each digit, using the carry out pin.

# 74393 Binary Counter
* The 74HC393 has two 4-bit binary counters, with a clock-in and reset pins for each counter. (MELEC p 217)

# Logic Chips
* Voltage input must be 5VDC regulated, with no spikes.
* Output of one chip can be connected to input of up to 10 others ("__fanout__").
* Inputs only sink about 1μA.
* Outputs can only provide up to 4mA.
* Wikipedia [7400-series integrated circuits](https://en.wikipedia.org/wiki/7400-series_integrated_circuits)
* __HC__ in the part number stands for [HCMOS](https://en.wikipedia.org/wiki/HCMOS) or "__high-speed CMOS__".
* Chips:
  * __AND__
    * Wikipedia [AND gate](https://en.wikipedia.org/wiki/AND_gate)
    * 7400: quad __2-input NAND__
    * 74HC08: quad __2-input AND__
    * 7410: triple __3-input NAND__
    * 7411: triple __3-input AND__
    * 7420: dual __4-input NAND__
    * 7430: single __8-input NAND__
    * 7421: dual __4-input AND__
  * __OR__
    * Wikipedia [OR gate](https://en.wikipedia.org/wiki/OR_gate)
    * 7402: quad __2-input NOR__
    * 74HC32: "quad __2-input OR__ gate (high speed CMOS version) - has lower current consumption/wider voltage range"
    * 7427: triple __3-input NOR__
    * 744075: triple __3-input OR__
    * 744002: dual __4-input NOR__
    * 744078: single __8-input OR and NOR__
  * __XOR__
    * 7486: quad __2-input XOR__
    * 747266: quad __2-input XNOR__
      * Wikipedia [XNOR gate](https://en.wikipedia.org/wiki/XNOR_gate)
  * __NOT__:
    * 7404 

# Flip-flops
* Wikipedia [Flip-flop](https://en.wikipedia.org/wiki/Flip-flop_(electronics)):
  "or latch is a circuit that has __two stable states__ and can be used to
  store state information – a bistable multivibrator. The circuit can be made
  to change state by signals applied to __one or more control inputs__ and will
  have __one or two outputs__. It is the basic storage element in sequential
  logic. Flip-flops and latches are fundamental building blocks of digital
  electronics systems used in computers, communications, and many other types
  of systems. ¶ Flip-flops and latches are used as data storage elements.
  A flip-flop is a device which __stores a single bit__ (binary digit) of data;
  one of its two states represents a "one" and the other represents a "zero"."
* The [555 timer](#555) can be used as a flip-flop.

# Debouncing
* Wikipedia [Switch: Contact bounce](https://en.wikipedia.org/wiki/Switch#Contact_bounce)
  "contact circuit voltages can be low-pass filtered to reduce or eliminate
  multiple pulses from appearing. In digital systems, multiple samples of the
  contact state can be taken at a low rate and examined for a steady sequence,
  so that contacts can settle before the contact level is considered reliable
  and acted upon. Bounce in SPDT switch contacts signals can be filtered out
  using a SR flip-flop (latch) or Schmitt trigger."
* A __decoupling capacitor__ is a capacitor used in combination with a resistor to smooth out
  the voltage change when a switch is pressed, so there's less need for debouncing. 
  He uses a 0.1uF capacitor in parallel with a pulldown resistor. (MELEC p 86)
* Later (MELEC p 212) he uses a 0.1uF capacitor again to debounce the input pin of a 555 timer, with the
  capacitor placed between the input pin and ground.
* Wikipedia [Decoupling capacitor](https://en.wikipedia.org/wiki/Decoupling_capacitor):
  "A decoupling capacitor is a capacitor used to decouple one part of an
  electrical network (circuit) from another. Noise caused by other circuit
  elements is shunted through the capacitor, reducing the effect it has on the
  rest of the circuit. For higher frequencies an alternative name is bypass
  capacitor as it is used to bypass the power supply or other high impedance
  component of a circuit."
* [555 timers](#555) can be used to debounce as well.
* A 4490 "bounce eliminator" chip is another option (e.g. the On Semiconductor MC14490, although it's relatively expensive
* __Jam-type flip-flops__ can be used to debounce. (MELEC p 214-216)
  * A SPDT switch is set to one of its two on positions and causes one of two
    corresponding outputs to go high while the other goes low.
  * These can be built with two NOR gates, or two NAND gates.
  * This doesn't work for a momentary switch, though.

# youtube.com [Collin's Lab: Electronics Tools](https://www.youtube.com/watch?v=Kv7Y8nAOoFE)
* Tweezers    
* Soldering iron: 
  * Xytronic XY-258
  * Weller WLC100: coil stand, cleaning sponge, cushy grip.
  * Metcal MX500: Rolls Royce. (ebay for cheap, or $500 US new.)
* Soldering station adjustable alligator grips
* Panavice Jr PCB holder
* Lupe magnifying glass.
* Magnifying visor.
* Multibit screwdriver set.

# Circuit Diagrams
* EasyEDA Electronic Circuit Design: Has library of components to design PCB layouts.
* youtube.com [From Idea to Schematic to PCB - How to do it easily!](https://www.youtube.com/watch?v=35YuILUlfGs)   

# youtube.com [Circuit Board Prototyping Tips and Tricks](https://www.youtube.com/watch?v=J9Ig1Sxhe8Y) (2016)
* Protoboards have 0.1" spacing because that's the spacing for [DIP IC](https://en.wikipedia.org/wiki/Dual_in-line_package) pins.
* Some project cases have slots that allow protoboards to be slid in (5:30)
* He likes protoboards with copper channels running length-wise. He uses a boarer tool to 
  move the copper anywhere he wants an insulation channel. (6:25)

# youtube.com [Control AC Devices with Arduino SAFELY - Relays & Solid State Switches](https://www.youtube.com/watch?v=H8FrL37Z7xE) (2020)
* Article: dronebotworkshop.com [Controlling AC Devices with Arduino](https://dronebotworkshop.com/ac-arduino/)
* Places prototype boards on mouse pad for insulation.
* Although for safety, he wires a step-down transformer to the 120V input to make it 28V.
* Also for safety, he uses an isolation transformer that takes the mains
  120V as input to produce a 120V output that is current limited, and probably
  has some type of breaker in it as well.
* He uses 3 separate boards: the Arduino, the relay module, and another board for the 28V lights
* Solid State Relays (SSR) are another option, instead of electromechanical relays. They require higher
  voltages on the AC side and slightly modify the AC wave form, but work fine for most applications.
  They have no moving parts, and so are good for faster switching circuits, and last longer.
* Wikipedia [Solid-state relay](https://en.wikipedia.org/wiki/Solid-state_relay)
* IoT Control Relays place all high voltage parts in a certified box for safety.
* He uses the [Digital Loggers IoT Relay](https://dlidirect.com/products/iot-power-relay)

# Home Assistant; 2022
* youtube.com [The new Raspberry Pi Pico W is just $6](https://www.youtube.com/watch?v=VEWdxvIphnI)
* Wikipedia [Home Assistant](https://en.wikipedia.org/wiki/Home_Assistant):
  "is a free and open-source software for home automation designed to be
  a central control system for smart home devices with a focus on local control
  and privacy."
* Wikipedia [Raspberry Pi](https://en.wikipedia.org/wiki/Raspberry_Pi)
  * Pico W: tiny, $6, has WiFi, released June 2022.
* Micropython

# QuadHands; 2022
* quadhands.com [QuadHands WorkBench Mount - Helping Hands Tool with Magnetic Arms & Panavise Mounting System](https://www.quadhands.com/collections/all/products/quadhands-workbench-mount) (2022)
* "Fits Any Panavise with the Three Point Mount"
* Cool Tools [QuadHands Helping Hands](https://www.youtube.com/watch?v=pN8HoQCmSjw): Uses the 
  [PANAVISE 301 2-1/2" Light Duty Multi-Angle Vise with Stationary Base K42480](https://www.amazon.com/gp/product/B07RWJFG7N?tag=seanmichaelragan-20&geniuslink=true)
* panavise.com [Model: 301  Standard PanaVise](https://www.panavise.com/index.html?pageID=1&page=full&--eqskudatarq=2):
  Will take any 5/8" diameter shaft.
* panavise.com [Model: 305  Low-Profile Base](https://www.panavise.com/index.html?pageID=1&page=full&--eqskudatarq=12):
  Is just a base.
* Looks like "Cool Tools" uses the [203](https://www.amazon.com/PanaVise-Vise-Head-Shaft-Bases/dp/B000SQWPY0/) attachment for soldering, with the 301.

