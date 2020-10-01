export const getMaxYpoint = (maxVal) => {
    let it = 1;
    while (true) {
        if (Math.pow(10, it) < maxVal) {
            it++;
        } else {
            break;
        }
    }
    const pow10 = Math.pow(10, it - 1)
    it = 2;
    while (true) {
        if (pow10 * it > maxVal) {
            return pow10 * it
        } else {
            it++;
        }
    }
}