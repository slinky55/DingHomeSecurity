#ifndef DING_ASYNC_CAM_H
#define DING_ASYNC_CAM_H

void getCameraStatus(AsyncWebServerRequest *request);
void setCameraVar(AsyncWebServerRequest *request);
void streamJpg(AsyncWebServerRequest *request);
void sendJpg(AsyncWebServerRequest *request);

#endif
