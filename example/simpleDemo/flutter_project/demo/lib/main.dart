import 'package:flutter/material.dart';

import 'package:flutter/foundation.dart' show debugDefaultTargetPlatformOverride;

import 'package:flutter/services.dart';
import 'dart:async';

void main() {
  // Desktop platforms aren't a valid platform.
  debugDefaultTargetPlatformOverride = TargetPlatform.fuchsia;
  runApp(new MyApp());
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return new MaterialApp(
      title: 'Flutter Demo',
      theme: new ThemeData(
        // If the host is missing some fonts, it can cause the
        // text to not be rendered or worse the app might crash.
        fontFamily: 'Roboto',
        primarySwatch: Colors.blue,
      ),
      home: new MyHomePage(title: 'Flutter Demo Home Page'),
    );
  }
}

class MyHomePage extends StatefulWidget {
  MyHomePage({Key key, this.title}) : super(key: key);

  final String title;

  @override
  _MyHomePageState createState() {
    return new _MyHomePageState();
  }
}

class _MyHomePageState extends State<MyHomePage> {
  static MethodChannel _channel = new MethodChannel('plugin_demo', new JSONMethodCodec());
  Future getVersion() async {
    var res = await _channel.invokeMethod('getNumber');
    print(res);
    setState(() {
      _counter = res;
    });
  }

  _MyHomePageState() {
    _channel.setMethodCallHandler((mc) async {
      switch (mc.method) {
        case "submit":
          final prev = _submittedMsg;
          setState(() {
            _submittedMsg = mc.arguments;
          });
          return prev;
      }
    });
    getVersion();
  }

  int _counter = 0;
  String _submittedMsg = "nothing yet";
  FocusNode myFocus = FocusNode();

  void _incrementCounter() {
    setState(() {
      _counter++;
    });
  }

  @override
  Widget build(BuildContext context) {
    return new Scaffold(
      appBar: new AppBar(
        title: new Text(widget.title),
      ),
      body: new Center(
        child: new Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            new Text(
              'You have pushed the button this many times:',
            ),
            new Text(
              '$_counter',
              style: Theme.of(context).textTheme.display1,
            ),
            new Text(
              'Last submit: ' + _submittedMsg,
            ),
            new Padding(
              padding: new EdgeInsets.all(8.0),
              child: new Column(children: <Widget>[
                TextField(
                  decoration: InputDecoration(hintText: 'TextField 1'),
                  onSubmitted: (value) {
                    setState(() {
                      _submittedMsg = value;
                    });
                    _channel.invokeMethod(
                      "print",
                      {
                        "textfield": value,
                        "number": _counter,
                      },
                    );
                  },
                  onEditingComplete: () => FocusScope.of(context).requestFocus(myFocus),
                ),
                TextField(
                  decoration: InputDecoration(hintText: 'TextField 2'),
                  maxLines: 2,
                  focusNode: myFocus,
                  onSubmitted: (value) {
                    setState(() {
                      _submittedMsg = value;
                    });
                  },
                ),
              ]),
            ),
          ],
        ),
      ),
      floatingActionButton: new FloatingActionButton(
        onPressed: _incrementCounter,
        tooltip: 'Increment',
        child: new Icon(Icons.add),
      ),
    );
  }
}
