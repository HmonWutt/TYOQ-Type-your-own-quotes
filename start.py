import curses
import time
from curses import color_pair, wrapper,textpad
text = '''The marmalade manufacturer discovered seventeen identical umbrellas
    beneath the railway station,'''
_='''each one precisely folded into a cube shape that 
    defied conventional geometry. Meanwhile, a retired accountant in Oslo was 
    simultaneously cataloging the migratory patterns of refrigerators—a pursuit 
    that had consumed the last four decades of his life with inexplicable passion. 
    The graffiti artist's masterpiece, rendered entirely in lowercase helvetica, 
    proclaimed that "socks dream of Antarctica," a statement that confused the city 
    council but resonated deeply with pigeons. Above it all, the automated espresso 
    machine at platform 3B continued its endless cycle of dispensing beverages to 
    passengers who existed only in theoretical discussions, their names preserved 
    in spreadsheets that nobody had opened since 2019.'''
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
    typed = ""
    win.addstr(text)
    continuous_score = curses.newwin(2,10,y+5,5)
    continuous_score.addstr(f"{len(typed)}/",curses.color_pair(YELLOW))
    continuous_score.addstr(f"{len(text_split)}")
    stdscr.refresh()
    win.refresh()
    continuous_score.refresh()
    curses.noecho()
    stdscr.nodelay(True)
    curses.curs_set(0)
    start = time.time()
    end = 0
    correct = 0
    typed_length = 0
    while len(typed) !=len(text):
        correct = 0
        try:
            c = win.getkey()
            if c == curses.KEY_BACKSPACE or c == '\x08' or c == '\x7f':
                typed = typed[:-1]
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
                break
            else:
                typed = typed+c
            win.erase()
            typed_length = len(typed)
            for i in range(typed_length):
                if typed[i] == text[i]:
                    correct+=1
                    win.addstr(text[i], curses.color_pair(GREEN))
                else:
                    win.addstr(text[i],curses.color_pair(RED))

            if typed_length<len(text):
                win.addstr(text[typed_length], UNDERLINE)
                if typed_length < len(text)-1:
                    win.addstr(text[typed_length+1:])
            else:
                end = time.time()
            win.refresh() 
            words_typed = typed.count(" ")
            continuous_score.erase()
            continuous_score.addstr(f"{words_typed}/",curses.color_pair(YELLOW))
            continuous_score.addstr(f"{len(text_split)}")
            continuous_score.refresh()

        except Exception as e:
            stdscr.addstr(str(e))          
            stdscr.refresh()
            pass
    if not len(typed) == len(text):
        return
    accuracy = (correct/typed_length)*100
    time_taken = end - start
    speed_character = (typed_length/time_taken)*60
    speed_word = (len(text)/5/time_taken)*60
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
    win.addstr("\nPress ESC key to exit the programme.", color_pair(RED))
    stdscr.refresh()
    win.refresh()
    stdscr.nodelay(False)
    while ord(c)!=27: 
        c = win.getkey()
def print_in_green(s,end): 
    print("\033[92m{}\033[00m".format(s),end=end,sep="",flush=True)
def print_in_purple(s, end): 
    print("\033[94m{}\033[00m".format(s),end=end,sep="",flush=True)
def get_input():
    instruction = "Paste your text here"
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
text = get_input()
wrapper(main,text)
