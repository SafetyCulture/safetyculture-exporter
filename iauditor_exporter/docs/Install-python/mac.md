# Install Python on a Mac

1. On the Mac, the best way to install Python is through a tool called Brew, so we'll install that first.
2. Visit: [https://brew.sh/](https://brew.sh/)
3. On the home page, you'll see a large "Install Homebrew" text, and below that a command that begins like this: `/usr/bin/ruby` 
4. Copy the entire command to your clipboard (highlight it all - if you double click it the entire line will be highlighted - then Right click > copy or press CMD+C)
    !['HomeBrew Website'](../../images/brew.png)
5. Open the macOS Terminal. Easiest way is to press the magnifying glass icon in the top right corner of your screen or press CMD+Space. From there, search for "Terminal" and press Enter.
    !['Spotlight Terminal'](../../images/spotlight-terminal.png)
6. In the Terminal window, paste the command (Edit > Paste, or CMD+V, or right click > paste) from before and again hit enter. 
7. Homebrew will explain what it intends to install, press Enter to let it go ahead. 
8. Once it's finished, you'll be ready to install Python. 
9. Run `brew install python3`
10. Assuming you don't get any errors, you're ready to use the script. If you get any errors, let us know. 

!!! Tip
    If `brew install python3` gives an error, try quitting your Terminal (CMD+Q) and opening it again. Sometimes it needs restarting to detect newly installed software. 