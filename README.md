gdp (GNOME Display Profiles)
============================

STATUS: EXPERIMENTAL / USE AT YOUR OWN RISK

Simple utility for saving and switching GNOME display configurations (i.e. profiles) under Wayland.

Install
-------

```
go build -o gdp
sudo install -m 0755 gdp /usr/local/bin/gdp
```

Usage
-----

1. Save current display configuration as `foo`.

   ```console
   $ gdp save foo
   ```

2. List display configurations

   ```console
   $ gdp list
   foo
   ```

3. Switch to `foo` profile

   ```console
   $ gdp switch foo
   $ gdp s foo # in short
   ```

Profiles are saved under `~/.config/gdp`.

Motivation
----------

I used to configure my displays using `xrandr` but it stopped working when I switched to Wayland. I needed an alternative solution.
