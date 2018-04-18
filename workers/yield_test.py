import inspect

def tryexcgenerator(func):
    def wrapper(*args, **kwargs):
        if inspect.isgeneratorfunction(func):
            try:
                for i in func(*args, **kwargs):
                    print('WRAP yield', i)
            except Exception as e:
                print("Exception", e)
        else:
            res = func(*args, **kwargs)
            print('WRAP return', res)
        return []
    return wrapper

@tryexcgenerator
def a():
    for i in range(10):
        yield i, 'AAA'
    # raise Exception('Hmm123')
    # yield StopIteration(999)
    return 999

for el in a():
    print(el)
