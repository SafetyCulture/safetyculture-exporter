## Setting up the script and installing requirements

In the previous steps you will have either downloaded the zip file from GitHub, or cloned the repository. Either of these options will have produced a folder containing the script. 

1. Open Windows Explorer (Windows) or Finder (macOS) and move to the folder where you downloaded the script. If you cloned it, it'll be called `iauditor_exporter`. If you downloaded the zip file the extracted folder will be called `iauditor_exporter_master` - if you're following along, you'll likely already have this open. 
2. Open your Command Prompt (Windows) or Terminal (macOS) - On Windows, click into the search box either on your Start Menu or on the taskbar, and type 'cmd' - press enter to launch the Command Prompt. On macOS, search for it using Spotlight (magnifying glass in the top right.)
3. In your terminal/cmd window, type: `cd` and press the spacebar once
4. Drag and drop either `iauditor_exporter` or `iauditor_exporter_master` onto the terminal/cmd window.
5. Press enter
6. The `cd` command means `change directory` so it tells the terminal/command prompt to move into that particular folder. 
7. If on Windows, run (copy and paste into your command prompt and hit enter): `pip install safetyculture-sdk-python/.`
8. If on macOS, run (copy and paste into your terminal and hit enter): `pip3 install safetyculture-sdk-python/.`
9. Once the above commands finish:
    * On Windows, run: `pip install -r requirements.txt`
    * On macOS, run: `pip3 install -r requirements.txt`
10. Assuming you do not get any errors, the script is now installed and you can continue to the configuration stage.
