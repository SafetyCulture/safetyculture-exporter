# Setting up the script and installing requirements

In the previous steps you will have either downloaded the zip file from GitHub, or cloned the repository. Either of these options will have produced a folder containing the script. 

## Steps for all users

1. Open Windows Explorer (Windows) or Finder (macOS) and move to the folder where you downloaded the script. If you cloned it, it'll be called `iauditor-exporter`. If you downloaded the zip file the extracted folder will be called `iauditor-exporter-master` - if you're following along, you'll likely already have this open. 

    !['Finder Folder'](../../images/folder-in-finder.png#thumbnail)
    
2. Open your Command Prompt (Windows) or Terminal (macOS) - On Windows, click into the search box either on your Start Menu or on the taskbar, and type 'cmd' - press enter to launch the Command Prompt. On macOS, search for it using Spotlight (magnifying glass in the top right.)

    !['Spotlight Terminal'](../../images/spotlight-terminal.png#thumbnail)

3. In your terminal/cmd window, type: `cd` and press the spacebar once
4. Drag and drop either `iauditor-exporter` or `iauditor-exporter-master` onto the terminal/cmd window then press Enter.
    
    !['Drag and drop finder'](../..w/images/drag-and-drop-finder.gif)
    
5. On Windows, run: `pip install -r requirements.txt`
6. On macOS, run: `pip3 install -r requirements.txt`
7. If you do not require SQL support, you can stop here. Otherwise, see below. 

## Additional Steps for SQL users

In addition to the requirements installed before, you'll also need an ODBC driver to interact with your SQL server.

1. Ensure you've completed the steps above first.
2. Visit Microsoft's website and [download the appropriate driver](https://docs.microsoft.com/en-us/sql/connect/odbc/download-odbc-driver-for-sql-server?view=sql-server-ver15)
3. On Windows, it's an installable exe file. Download it and Windows will take you through the installation. On macOS, you'll be redirected to a series of commands to run in your terminal. Run them one line at a time, in the same way you have for the other steps in this documentation. 


!!! Tip
    The `cd` command means `change directory` so it tells the terminal/command prompt to move into that particular folder. You can run `ls` to see the contents of the folder, too. 