void i444ToI420CGO(
    unsigned char* cb,
    unsigned char* cr,
    const int stride, const int h);

void i422ToI420CGO(
    unsigned char* cb,
    unsigned char* cr,
    const int stride, const int h);

void rgbToYCbCrCGO(
    unsigned char* y,
    unsigned char* cb,
    unsigned char* cr,
    const unsigned char r,
    const unsigned char g,
    const unsigned char b);  // for testing

void repeatRGBToYCbCrCGO(
    const int n,
    unsigned char* y,
    unsigned char* cb,
    unsigned char* cr,
    const unsigned char r,
    const unsigned char g,
    const unsigned char b);  // for testing

void yCbCrToRGBCGO(
    unsigned char* r,
    unsigned char* g,
    unsigned char* b,
    const unsigned char y,
    const unsigned char cb,
    const unsigned char cr);  // for testing

void repeatYCbCrToRGBCGO(
    const int n,
    unsigned char* r,
    unsigned char* g,
    unsigned char* b,
    const unsigned char y,
    const unsigned char cb,
    const unsigned char cr);  // for testing

void i444ToRGBACGO(
    unsigned char* rgb,
    const unsigned char* y,
    const unsigned char* cb,
    const unsigned char* cr,
    const int stride, const int h);

void rgbaToI444(
    unsigned char* y,
    unsigned char* cb,
    unsigned char* cr,
    const unsigned char* rgb,
    const int stride, const int h);
