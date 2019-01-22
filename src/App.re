open Belt;

let url = "http://localhost:5000/";

type heaterStatus = {
  minTemp: float,
  maxTemp: float,
  lastTempReading: float,
  econoMode: bool,
  disabled: bool,
  forcedOnTimeLimit: option(Js.Date.t),
  heaterOn: bool,
};

type state =
  | Unknown
  | Known(heaterStatus);

type action =
  | GetStatus
  | GotStatus(heaterStatus)
  | Disable
  | Enable;

let component = ReasonReact.reducerComponent(__MODULE__);

module Api = {
  let decode = (j: Js.Json.t): heaterStatus => {
    lastTempReading:
      j |> Json.Decode.field("last_temp_reading", Json.Decode.float),
    heaterOn: j |> Json.Decode.field("heater_on", Json.Decode.bool),
    minTemp: j |> Json.Decode.field("min_temp", Json.Decode.float),
    maxTemp: j |> Json.Decode.field("max_temp", Json.Decode.float),
    econoMode: j |> Json.Decode.field("econo_mode", Json.Decode.bool),
    disabled: j |> Json.Decode.field("disabled", Json.Decode.bool),
    forcedOnTimeLimit:
      j
      |> Json.Decode.optional(
           Json.Decode.field("forced_on_time_limit", Json.Decode.date),
         ),
  };

  let enable = () =>
    Js.Promise.(
      Fetch.fetchWithInit(
        url ++ "disable",
        Fetch.RequestInit.make(~method_=Delete, ()),
      )
      |> then_(Fetch.Response.json)
      |> then_(json => decode(json) |> resolve)
    );

  let disable = () =>
    Js.Promise.(
      Fetch.fetchWithInit(
        url ++ "disable",
        Fetch.RequestInit.make(~method_=Post, ()),
      )
      |> then_(Fetch.Response.json)
      |> then_(json => decode(json) |> resolve)
    );

  let getStatus = () =>
    Js.Promise.(
      Fetch.fetch(url ++ "status")
      |> then_(Fetch.Response.json)
      |> then_(json => decode(json) |> resolve)
    );
};

let make = _children => {
  ...component,
  initialState: () => Unknown,
  reducer: (action, state: state) => {
    switch (action) {
    | Enable =>
      ReasonReact.SideEffects(
        self => {
          Js.log("side effect enable called");
          Api.enable()
          |> Js.Promise.then_(heaterStatus =>
               self.send(GotStatus(heaterStatus)) |> Js.Promise.resolve
             );
          ();
        },
      )
    | Disable =>
      ReasonReact.SideEffects(
        self => {
          Js.log("side effect disable called");
          Api.disable()
          |> Js.Promise.then_(heaterStatus =>
               self.send(GotStatus(heaterStatus)) |> Js.Promise.resolve
             );
          ();
        },
      )
    | GotStatus(htrSts) => ReasonReact.Update(Known(htrSts))
    | GetStatus =>
      ReasonReact.SideEffects(
        self => {
          Api.getStatus()
          |> Js.Promise.then_(heaterStatus =>
               self.send(GotStatus(heaterStatus)) |> Js.Promise.resolve
             );
          ();
        },
      )
    };
  },
  didMount: self => {
    self.send(GetStatus);
  },
  render: self => {
    let toggle = currentlyDisabled => {
      Js.log2("value of currently Enabled", currentlyDisabled);
      if (currentlyDisabled) {
        self.send(Enable);
      } else {
        self.send(Disable);
      };
      ();
    };
    switch (self.state) {
    | Unknown => <div> {ReasonReact.string("Unknown")} </div>
    | Known(hs) =>
      <div>
        <h1>
          {ReasonReact.string(hs.heaterOn ? "Heater is ON" : "Heater is OFF")}
        </h1>
        <h2>
          {ReasonReact.string(
             "Temperature: " ++ string_of_float(hs.lastTempReading),
           )}
        </h2>
        <h3>
          {ReasonReact.string(
             "Heater will turn on when temp is below: "
             ++ string_of_float(hs.minTemp),
           )}
        </h3>
        <h3>
          {ReasonReact.string(
             "Heater will turn off when temp is above: "
             ++ string_of_float(hs.maxTemp),
           )}
        </h3>
        <label className="switch">
          <input
            type_="checkbox"
            checked={!hs.disabled}
            onClick={_ => toggle(hs.disabled)}
          />
          <span className="slider round" />
        </label>
        {hs.disabled ?
           <div> {ReasonReact.string("The heater is at ")} </div> :
           ReasonReact.null}
      </div>
    };
  },
};

/*


 type action =
   | LoadInvoicesRequested
   | LoadInvoicesSucceeded(array(Invoice.t))
   | LoadInvoicesFailed;

 module Decode = {
   let invoices = json: array(Invoice.t) =>
     Json.Decode.array(Invoice.decoder, json);
 };

 let fetchInvoices = (): Js.Promise.t(option(array(Invoice.t))) =>
   Js.Promise.(
     Fetch.fetch(url)
     |> then_(Fetch.Response.json)
     |> then_(json =>
          json |> Decode.invoices |> (invoices => Some(invoices) |> resolve)
        )
     |> catch(err => {
          Js.log2("in catch", err);
          resolve(None);
        })
   );

 let processInvoiceResults = (self, results) =>
   Js.Promise.(
     results
     |> then_(result =>
          switch (result) {
          | Some(invoices) =>
            resolve(self.ReasonReact.send(LoadInvoicesSucceeded(invoices)))
          | None => resolve(self.send(LoadInvoicesFailed))
          }
        )
     |> ignore
   );

 let component = ReasonReact.reducerComponent(__MODULE__);

 let showApp = state =>
   switch (state) {
   | NotAsked => <div> {ReasonReact.string("Initializing...")} </div>
   | Loading => <div> {ReasonReact.string("Loading invoices")} </div>
   | Failure => <div> {ReasonReact.string("Failed to load invoices")} </div>
   | Success(invoices) => <div> <InvoicesComponent invoices /> </div>
   };

 let make = _children => {
   ...component,
   initialState: () => NotAsked,
   didMount: self => self.send(LoadInvoicesRequested),
   reducer: (action, _state) =>
     switch (action) {
     | LoadInvoicesRequested =>
       ReasonReact.UpdateWithSideEffects(
         Loading,
         (self => fetchInvoices() |> processInvoiceResults(self)),
       )
     | LoadInvoicesSucceeded(invoices) =>
       ReasonReact.Update(Success(invoices))
     | LoadInvoicesFailed => ReasonReact.Update(Failure)
     },
   render: self => showApp(self.state),
 }; */