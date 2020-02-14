# Installation

## Overview

If you're comfortable using Git and running commands from the command prompt, these instructions should give you what you need:

* Install Python 3.6+
* If you've used the SafetyCulture Python SDK before, you need to uninstall it (if you're unsure, run the command anyway): `pip uninstall safetyculture-sdk-python`
* Clone or download the repository and move to it in terminal/command prompt
* Run `pip install safetyculture-sdk-python/.`
* Run `pip install -r requirements.txt`
* Open `configs/config.yaml` and edit as needed
* Run `python exporter.py --format sql` (Or other format support by the exporter)

## Detailed Guidance

If you're new to running scripts, follow the guidance below.

## Install Python on Windows

1. The first thing you need is Python itself. Python is available in many different versions, ranging from version 2.7 all the way through to 3.8 at the time of writing. For our script, you'll need at least Python 3.6
2. Visit Python's Windows Releases Page: https://www.python.org/downloads/windows/
3. This page contains a lot of links, the section you need is on the left hand side titled "Stable Releases"
4. Under "Stable Releases" you'll see various download options. In 99% of cases you'll be able to use the one called "Windows x86-64 exectuable installer"
5. This will begin the download, once done open the file and click through the installer - the defaults are fine, feel free to keep clicking continue through them. 
6. On the final page before clicking 'Finish' there is a tick box that reads "Add to PATH" - ensure this is ticked. If you're reading this having already missed the tickbox, simply re-run the installer, and it'll ask again at the end. 
7. That's it! We can now move on to the script. 

## Install Python on a Mac

1. On the Mac, the best way to install Python is through a tool called Brew, so we'll install that first.
2. Visit: https://brew.sh/
3. On the home page, you'll see a large "Install Homebrew" text, and below that a command that begins like this: `/usr/bin/ruby` 
4. Copy the entire command to your clipboard (highlight it all - if you double click it the entire line will be highlighted - then Right click > copy or press CMD+C)
5. Open the macOS Terminal. Easiest way is to press the magnifying glass icon in the top right corner of your screen or press CMD+Space. From there, search for "Terminal" and press Enter.
6. In the Terminal window, paste the command (Edit > Paste, or CMD+V, or right click > paste) from before and again hit enter. 
7. Homebrew will explain what it intends to install, press Enter to let it go ahead. 
8. Once it's finished, you'll be ready to install Python. Run `brew install python3`
9. Assuming you don't get any errors, you're ready to use the script. If you get any errors, let us know. 

