#include "base64_encode.h"
#include <fstream>
#include <iterator>
#include <stdexcept>
#include <string>

std::string base64_encode(const std::string& in) {
    const std::string chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
    std::string out;
    int val = 0, valb = -6;
    for (unsigned char c : in) {
        val = (val << 8) + c;
        valb += 8;
        while (valb >= 0) {
            out.push_back(chars[(val >> valb) & 0x3F]);
            valb -= 6;
        }
    }
    if (valb > -6) out.push_back(chars[((val << 8) >> (valb + 8)) & 0x3F]);
    while (out.size() % 4) out.push_back('=');
    return out;
}

std::string file_to_base64(const std::string& path) {
    std::ifstream f(path, std::ios::binary | std::ios::ate); 
    if (!f) throw std::runtime_error("file not found: " + path);

    std::streamsize size = f.tellg();
    f.seekg(0, std::ios::beg);

    std::string buf(size, '\0');
    if (f.read(&buf[0], size)) {
        return base64_encode(buf);
    }
    return "";
}
