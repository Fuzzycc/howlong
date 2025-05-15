# howlong
A quick tool to tell you how long downloading X size will take with Z speed.

# The Inception
Due to a slow internet and the desire to download large files (read: video games), I always find myself opening a calculator and going:

120 gigabytes.... times 1024 to megabytes... times 8 to megabits... divide by 3 megabits per second... divide by 3600 to hours... 91 hours, sigh... divide by 24 to days... 3.7 days of cumulative downloads...

And so I finally faced myself in the mirror, realized I'm a programmer, and decided to write a simple tool to make life easier and further smoothen the brain.

Even the Wizards created cantrips to make life easier, so why can't I?

# How does it work
The complete command takes a size string composed of a number and a unit, 30GB for example, a speed string of the same format, and a time unit.

The command then returns the Estimated Time to Download or ETD.

# Continuous Mode
Continuous mode is started with the `--continuous` or `-c` flag and supplied with `{unit} {unit} {time-unit}`
Then, input is taken in the form of `{down-size} {speed}` and an output is given in the form of `duration` concatted with `time-unit`

## Example
```
$hl -c GB KB h
>10 420 
6.9350h
>10 369
7.8933h
-1
```
The program will output -1 when terminated.

# Help
hl {down-size[unit]} {speed[unit]} [time-unit]

down-size and speed are a float with decimal precision of 2

unit uses B for Byte and b for bit:
- GB (down-size Default)
- Gb
- MB
- Mb (speed Default)
- KB
- Kb
- B
- b

unit given to speed are treated on a per-second basis for now. (384KB means 384 KiloBytes per second of speed)

time-unit has a default of hours with a decimal precision of 2 for the output:
- d
- h
- m
- s

# Example
```
$hl 10 3
7.58
```
This uses the configured unit of GB, Mbps, and hours output.
```
$hl 120GB 3Mb d
3.79
```
This explicitly asks for GigaBytes, Megabits per second, and sets the output to days.

# Next
What's next to first release is incorporating configuration to allow pre-setting defaults to increase efficiency and ease of use

# Afterwords
The wizard has 104 GigaMana and a casting speed factor of 10 MMps.
The Fireball has already been cast and I have been utterly annihi---
