# Installing Git

We host the script on GitHub, so the easiest way to download the script is by what's known as 'cloning' the repository. Cloning has other additional benefits, mainly around updating the code as and when we publish updates.

## Downloading Git on Windows

!!! Warning
    Please note you'll need to be an admin to install Git successfully. If you're not, either speak to your administrator or use the [manual process](../../script-download/manual-dl)

1. You can download Git for Windows from this link - the download will start automatically (if it doesn't, click 'Click here to download manually'): [https://git-scm.com/download/win](https://git-scm.com/download/win)
2. Open the installer and begin installation. All the defaults are fine so keep clicking continue until it begins installing. Once done, click 'Next' again to close the installer (the installer offers to launch Git but this isn't necessary for our purposes.)
4. Assuming you do not get any errors, you can continue to download the script. Keep your command prompt open and follow the 'Download the Script' steps below.

## Download Git on macOS
1. If you've been following this guidance from the beginning, you'll have installed Homebrew to install Python. If you didn't, [go back and install Homebrew](../../Install-python/mac/). 
2. Open your terminal window again (remember you can get to it by searching for it by clicking the magnifying glass in the top right)
3. Run `brew install git`
4. Assuming you do not get any errors, you can continue to download the script. Keep your terminal window open for the next step.

## Download the Script
1. In your terminal or command prompt, we need to move to the location where you want to store the script. We recommend your Documents folder.
2. Navigate to the folder where you want to store the script in either Finder on macOS or Windows Explorer if you're on Windows.
2. In your terminal or command prompt type `cd` and then press the spacebar once (but don't press enter)
3. Drag and drop the folder from Finder/Windows Explorer, and drop it into the terminal/command prompt. Press Enter.

   !['Drag and drop finder'](../../images/drag-and-drop-finder.gif)
   
4. You can now run `git clone https://github.com/SafetyCulture/iauditor-exporter.git`
5. Once completed, we can move on to the configuration stage.  
        
        
