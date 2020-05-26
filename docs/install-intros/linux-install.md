# Notes on using Linux

Setting up the tool on Linux is very similar to macOS. Most Linux users will be able to use the [quick install](../../install-intros/quick-install) guide to get the initial set up done. 

In testing using Ubuntu 18 LTS, the only additional requirements required on a fresh install were:

 * Git (`sudo apt install git`)
 * Pip for Python3 (`sudo apt install python3-pip`)
 * Unix ODBC Library (`sudo apt install unixodbc-dev`) 
 
Of course if you're using a distro other than Ubuntu, you'll need to use the relevant package manager. 

If you experience other dependency requirements on your flavour of Linux, please raise an issue with the solution and I'll append it to this page so others can benefit.

# Raspberry Pi

There have been some queries around running this tool on a Raspberry Pi (running Raspbian.) It would appear that Numpy can have some issues on a Pi, giving an error:

``` 
import error: libf77blas.so.3: cannot open shared object file:No such file or directory
``` 

This is likely due to missing dependencies. Try:

```
sudo apt-get install python-dev libatlas-base-dev
```

You may need to re-run `pip install -r requirements.txt` afterwards. 

Further Reading: https://stackoverflow.com/questions/53784520/numpy-import-error-python3-on-raspberry-pi