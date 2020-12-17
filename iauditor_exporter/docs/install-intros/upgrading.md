# Updating

## Backup first

Regardless of how you run the update, it's imperative that you make a back up first as to avoid possible data loss. 

Inside the `iauditor-exporter` folder are many other files and folders. The ones you should backup are:

* configs
* last_successful
* exports

Simply copy the three folders and paste them somewhere else outside of the `iauditor-exporter` folder, such as on your Desktop. 
]
## If you used Git

If you used Git, you can `cd` to your `iauditor-exporter` folder and run `git pull` to update. Check out the install instructions if you've forgotten how to do this. 

## If you manually downloaded the zip file 

* Download the script again: https://github.com/SafetyCulture/iauditor-exporter/archive/master.zip
* Extract the zip file
* Drag and drop the contents of the folder `iauditor-exporter-master` into your existing script, allowing it to replace any files that already exist. 

## Troubleshooting

Should you receive errors after upgrading, the simplest solution is to do a fresh install. Once installed, you can copy and paste your backed up `configs`, `last_successful`, and `exports` folder back in. 