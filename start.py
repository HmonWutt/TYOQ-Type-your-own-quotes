import curses
from curses import wrapper
import time
def main(stdscr):

    curses.init_pair(1, curses.COLOR_RED, curses.COLOR_WHITE)
    # Clear screen
    #stdscr.clear()

    # This raises ZeroDivisionError when i == 10.
#    for i in range(0, 10):
#    v = i-10
#        stdscr.addstr(i, 10, '10 divided by {} is {}'.format(v, 10/v), curses.color_pair(1))
#        stdscr.refresh() 
#        stdscr.getkey()
    stdscr.nodelay(True)
    c = ""
    stdscr.addstr("hello welcome to the world's biggest typing championship")
    win = curses.newwin(10,10,1,0)
    stdscr.nodelay(True)
    curses.curs_set(2)
    y,x =win.getmaxyx()
    stdscr.addstr(str(y) +","+str(x))
    while True:
        curses.flash()
        try:
            #c = stdscr.getkey()
            #stdscr.addstr(10,10,c,curses.A_UNDERLINE | curses.A_BOLD|curses.color_pair(1))
            #stdscr.refresh()
            c = stdscr.getkey()
            if c == 'q': 
                break

            elif c == 'KEY_BACKSPACE' or c == '\x08' or c == '\x7f':
                y,x = win.getyx()

                if x ==0 :

                    if y >0:
                        y-=1
                        x = 9

                    win.move(y,x)
                    win.delch(y,x)

                
                win.delch(y,x-1)
            else:
                win.addstr(c,curses.A_BOLD | curses.A_UNDERLINE| curses.color_pair(1))
            
            
                y,x = win.getyx()

            #stdscr.addstr(str(y) +","+str(x)+"\n")
            win.refresh()


        except curses.error:
            #stdscr.addstr("Error")
            pass

wrapper(main)
