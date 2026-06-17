import curses
from curses import wrapper
def main(stdscr):

    RED =1 
    GREEN = 2
    UNDERLINE = curses.A_UNDERLINE
    curses.init_pair(RED, curses.COLOR_RED, curses.COLOR_BLACK)
    curses.init_pair(GREEN, curses.COLOR_GREEN, curses.COLOR_BLACK)

    c = ""
    header = "Let's see how fast you can type!"
    height,width= stdscr.getmaxyx()
    text_width = len(header) + 4
    x = (width - text_width) // 2
    y = 2
    stdscr.addstr(y - 1, x, "+" + "-" * (text_width) + "+")
    stdscr.addstr(y, x, "|" + " " * (text_width) + "|")
    stdscr.addstr(y + 1, x, "|  " + header + "  |")
    stdscr.addstr(y + 2, x, "|" + " " * (text_width) + "|")
    stdscr.addstr(y + 3, x, "+" + "-" * (text_width) + "+")
    

    win = curses.newwin(height,width-10,y+5,5)
    #text = "This is the text to type."
    text = '''The marmalade manufacturer discovered seventeen identical umbrellas
    beneath the railway station, each one precisely folded into a cube shape that 
    defied conventional geometry. Meanwhile, a retired accountant in Oslo was 
    simultaneously cataloging the migratory patterns of refrigerators—a pursuit 
    that had consumed the last four decades of his life with inexplicable passion. 
    The graffiti artist's masterpiece, rendered entirely in lowercase helvetica, 
    proclaimed that "socks dream of Antarctica," a statement that confused the city 
    council but resonated deeply with pigeons. Above it all, the automated espresso 
    machine at platform 3B continued its endless cycle of dispensing beverages to 
    passengers who existed only in theoretical discussions, their names preserved 
    in spreadsheets that nobody had opened since 2019.'''
    text = ' '.join(text.split())
    win.addstr(text)

    stdscr.refresh()
    win.refresh()
    curses.noecho()
    stdscr.nodelay(True)
    curses.curs_set(0)
    typed = ""
    while typed !=text:
        try:
            c = win.getkey()
            if c == 'q': 
                break
            elif c == curses.KEY_BACKSPACE or c == '\x08' or c == '\x7f':
                typed = typed[:-1]
            elif c =="KEY_RESIZE":
                height,width= stdscr.getmaxyx()
                win.resize(height,width-10)
                stdscr.erase()
                text_width = len(header) + 4
                x = (width - text_width) // 2
                y = 2
                stdscr.addstr(y - 1, x, "+" + "-" * (text_width) + "+")
                stdscr.addstr(y, x, "|" + " " * (text_width) + "|")
                stdscr.addstr(y + 1, x, "|  " + header + "  |")
                stdscr.addstr(y + 2, x, "|" + " " * (text_width) + "|")
                stdscr.addstr(y + 3, x, "+" + "-" * (text_width) + "+")
                stdscr.refresh()

            elif c == curses.KEY_MOUSE:
                curses.flash()
                pass
            else:
                typed = typed+c
            win.erase()
            limit = len(typed)
            if limit == len(text):
                break
            for i in range(limit):
                if typed[i] == text[i]:
                    win.addstr(typed[i], curses.color_pair(GREEN))
                else:
                    win.addstr(typed[i],curses.color_pair(RED))
            if limit<len(text):
                win.addstr(text[limit], UNDERLINE)
                win.addstr(text[limit+1:])
            win.refresh()

        except curses.error:
            pass

wrapper(main)
