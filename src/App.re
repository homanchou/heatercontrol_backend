open Belt;

let url = "http://localhost:5000/status";

type heaterStatus = {
  minTemp: float,
  maxTemp: float,
  econoMode: bool,
  disabled: bool,
  forcedOnTimeLimit: option(Js.Date.t),
};

type state =
  | Unknown
  | Known(heaterStatus);

type action =
  | GetStatus
  | GotStatus(heaterStatus)
  | Disable;

let component = ReasonReact.reducerComponent(__MODULE__);

module Api = {
  let decode = (j: Js.Json.t): heaterStatus => {
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

  let getStatus = () =>
    Js.Promise.(
      Fetch.fetch(url)
      |> then_(Fetch.Response.json)
      |> then_(json => decode(json) |> resolve)
    );
};

let make = _children => {
  ...component,
  initialState: () => Unknown,
  reducer: (action, state: state) => {
    switch (action) {
    | Disable =>
      ReasonReact.SideEffects(self => Js.log("side effect disable called"))
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
    switch (self.state) {
    | Unknown => <div> {ReasonReact.string("Unknown")} </div>
    | Known(hs) =>
      <div>
        <h1> {ReasonReact.string("Enabled")} </h1>
        <label className="switch">
          <input
            type_="checkbox"
            checked={!hs.disabled}
            onClick={_ => self.send(Disable)}
          />
          <span className="slider round" />
        </label>
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