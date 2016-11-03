#-*- coding: utf-8 -*-   

# Thanks to @superalsrk 

# The encode may be gb18030 to varing operation system
# Download ActivePython-3.4.3.2-win64-x64.msi and install
# Use pypm to install pyautogui and requests
# Run :  python win_client.py

import win32gui
import win32api
import win32con
import time
import ctypes
import requests
import pyautogui
import datetime

print(u'Start..')
counter = 1 

## fetch the mouse position by  pyautogui.position()  
first_ret = (666, 666)
rel_link = (888, 464)


## 通过查看历史文章的页面链接获取到对应的bizid
bizs = ["MzA5NzEyMTQ2NQ==", "MzIzMDE0MDQyNQ==", "MjM5MDEyMDk4Mw=="]

def gen():
   global counter
   return bizs[counter % len(bizs)]

def process():
   global counter
   print('Requests Number:  %d  Biz (%s)\n----------------------------------------------------' % (counter,  datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')))
   
   print(pyautogui.size())
   print(pyautogui.position())
   
   ## 获取句柄
   hwnd = win32gui.FindWindow(None, u"微信")
   win32gui.SetForegroundWindow(hwnd)
   win32gui.SetForegroundWindow(hwnd)
   
   pyautogui.moveTo(x=first_ret[0], y=first_ret[1], duration=1)
   pyautogui.click(x=first_ret[0], y = first_ret[1])

   biz = gen()

   pyautogui.typewrite(" http://mp.weixin.qq.com/mp/getmasssendmsg?__biz=%s==#wechat_webview_type=1&wechat_redirect" % biz, interval=0)

   pyautogui.keyDown('enter')
   pyautogui.moveTo(rel_link[0], rel_link[1], duration=1)
   pyautogui.click(x=rel_link[0], y=rel_link[1])
   counter = counter + 1
   
   
if __name__ == '__main__':
   while True:
      process()
      time.sleep(3)