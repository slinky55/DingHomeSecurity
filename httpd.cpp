#include <SPIFFS.h>

#include <AsyncEventSource.h>
#include <AsyncJson.h>
#include <AsyncWebSocket.h>
#include <AsyncWebSynchronization.h>
#include <ESPAsyncWebServer.h>
#include <SPIFFSEditor.h>
#include <StringArray.h>
#include <WebAuthentication.h>
#include <WebHandlerImpl.h>
#include <WebResponseImpl.h>

AsyncWebServer* apServer = NULL;

AsyncWebServer* streamServer = NULL;

extern bool connectedToBase;

void startApServer() 
{ 
  if (streamServer) {
    Serial.println("Stream server still running, can't start AP server...");
    return;
  }
  apServer = new AsyncWebServer(80);

  DefaultHeaders::Instance().addHeader("Access-Control-Allow-Origin", "*");

  apServer->serveStatic("/", SPIFFS, "/res");

  apServer->on("/", HTTP_GET, [](AsyncWebServerRequest* req) {
    req->send(SPIFFS, "/ap.html");
  });

  apServer->addHandler(new AsyncCallbackJsonWebHandler("/connect", [](AsyncWebServerRequest *request, JsonVariant &json) {
    const char* id = json["id"];
    const char* pass = json["pass"];
    Serial.println(id);
    Serial.println(pass);
  }));

  apServer->begin();
}

void stopApServer() 
{
  apServer->end();
  delete apServer;
}

void startCameraServer() 
{
  if (apServer) {
    Serial.println("AP server still running, can't start stream server...");
    return;
  }
  streamServer = new AsyncWebServer(80);
}

void stopCameraServer()
{
  delete streamServer;
}
