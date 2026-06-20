import curses
import time
import sys
import random
from curses import color_pair, wrapper

def main(stdscr,text):

    RED =1 
    GREEN = 2
    YELLOW = 3
    BLUE = 4
    MAGNETA = 5
    BLACK = curses.COLOR_BLACK
    UNDERLINE = curses.A_UNDERLINE
    curses.init_pair(RED, curses.COLOR_RED, BLACK)
    curses.init_pair(GREEN, curses.COLOR_GREEN, BLACK)
    curses.init_pair(YELLOW, curses.COLOR_YELLOW, BLACK)
    curses.init_pair(BLUE, curses.COLOR_BLUE, BLACK)
    curses.init_pair(MAGNETA, curses.COLOR_MAGENTA, BLACK)

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
    
    win = curses.newwin(height,width-10,y+7,5)
    #text = "This is the text to type."
 
    text_split  = text.split()
    text = ' '.join(text_split)
    win.addstr(text)
    continuous_score = curses.newwin(2,10,y+5,5)
    continuous_score.addstr(f"0/",curses.color_pair(YELLOW))
    continuous_score.addstr(f"{len(text_split)}")
    stdscr.refresh()
    win.refresh()
    continuous_score.refresh()
    curses.noecho()
    stdscr.nodelay(True)
    curses.curs_set(0)
    start = time.time()
    end = 0
    error_indxs = []
    indx  = -1
    while indx !=len(text)-1:
        try:
            c = win.getkey()
            if c == curses.KEY_BACKSPACE or c == '\x08' or c == '\x7f':
                if error_indxs:
                    last_error_index = error_indxs[-1]
                    if indx == last_error_index:
                        error_indxs.pop()

                if indx >-1:
                    indx-=1
            elif c =="KEY_RESIZE" or c == curses.KEY_RESIZE:
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

            elif ord(c) == 27: 
                return "done"
            elif c == text[indx+1]:
                indx +=1
            else:
                indx +=1
                error_indxs.append(indx)

            win.erase()
            for i in range(indx+1):
                if i not in error_indxs:
                    win.addstr(text[i], curses.color_pair(GREEN))
                else:
                    win.addstr(text[i],curses.color_pair(RED))

            if indx < len(text)-1:
                win.addstr(text[indx+1], UNDERLINE)
                if indx < len(text)-2:
                    win.addstr(text[indx+2:])
            else:
                end = time.time()
            win.refresh() 
            words_typed = text[:indx+1].count(" ")
            continuous_score.erase()
            continuous_score.addstr(f"{words_typed}/",curses.color_pair(YELLOW))
            continuous_score.addstr(f"{len(text_split)}")
            continuous_score.refresh()

        except Exception as e:
            stdscr.addstr(str(e))          
            stdscr.refresh()
            pass
    if indx < len(text)-1:
        return
    total = len(text)
    errors = len(error_indxs)
    correct = total - errors
    accuracy = (correct/total)*100
    time_taken = end - start
    speed_character = (total/time_taken)*60
    speed_word = (total/5/time_taken)*60
    win.clear()
    stdscr.clear()
    header = ("(⌐■_■) These are your results")    
    height,width= stdscr.getmaxyx()
    text_width = len(header) + 4
    x = (width - text_width) // 2
    y = 2
    stdscr.addstr(y - 1, x, "+" + "-" * (text_width) + "+")
    stdscr.addstr(y, x, "|" + " " * (text_width) + "|")
    stdscr.addstr(y + 1, x, "|  " + header + "  |")
    stdscr.addstr(y + 2, x, "|" + " " * (text_width) + "|")
    stdscr.addstr(y + 3, x, "+" + "-" * (text_width) + "+")
    win.addstr(f"Speed: {round(speed_word)} wpm\n", curses.color_pair(BLUE))
    win.addstr(f"Speed: {round(speed_character)} cpm\n",curses.color_pair(MAGNETA))
    win.addstr(f"accuracy: {round(accuracy)} %",curses.color_pair(YELLOW))
    win.addstr("\nPress ESC key to exit the programme. Press any other key to type again", color_pair(RED))
    stdscr.refresh()
    win.refresh()
    stdscr.nodelay(False)
    while True: 
        c = win.getkey()
        if ord(c)==27:
            return "done"
        return "continue"
def print_in_green(s,end): 
    print("\033[92m{}\033[00m".format(s),end=end,sep="",flush=True)

def print_in_purple(s, end): 
    print("\033[94m{}\033[00m".format(s),end=end,sep="",flush=True)

def get_input():
    instruction = "Welcome to TYOQ. Paste your text below"
    length = len(instruction)+4
    h_line = "+"+"-"* length+"+"
    v_line = "|"+ " "*length +"|"
    v_line_mid = "|  "+instruction+"  |"
    formatted_instruction = h_line+"\n"+v_line+"\n"+v_line_mid+"\n"+v_line+"\n"+h_line
    print_in_green(formatted_instruction,"\n")
    inpt = input()
    print_dots()
    return inpt

def print_dots():
    count = 4
    print_in_green("Input saved. Redirecting to the typing arena ","")
    for _ in range(count):
        print_in_green("➤","")
        time.sleep(0.5)

args = sys.argv
quote1 = """
It is often better to light a flamethrower than curse 
the darkness, although the results are much the same.
"""

quote2 = """
The pen is mightier than the sword if the sword is very 
short and the pen is very sharp.
"""

quote3 = """
Build a man a fire, and he'll be warm for a day. Set a man 
on fire, and he'll be warm for the rest of his life.
"""

quote4 = """
It is said that your life flashes before your eyes just 
before you die. That is true, it's called Living.
"""

quote5 = """
Always remember that the crowd that applauds your coronation 
is the same crowd that will applaud your beheading.
"""
ipt = ""
quotes = [quote1,quote2,quote3,quote4]
if __name__ == '__main__':
    if len(sys.argv)<2:
        while True:
            pick = random.randint(0,len(quotes)-1)
            text = quotes[pick]
            ipt = wrapper(main,text)
            if ipt == "done":
                break
    else:
        while True:
            text = get_input()
            ipt =wrapper(main,text)
            if ipt=="done":
                break

