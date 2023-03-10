module Main exposing (main)

import Browser
import Html exposing (..)
import Html.Attributes exposing (attribute, class, placeholder, required, type_, value)
import Html.Events exposing (onInput, onSubmit)
import Http
import Json.Encode as E


main : Program () Model Msg
main =
    Browser.element
        { init = init
        , view = view
        , update = update
        , subscriptions = \_ -> Sub.none
        }



-- MODEL


type alias Model =
    { form : Form
    , page : Page
    }


type alias Form =
    { name : String
    , child : String
    , amount : Int
    }


type Page
    = MainPage
    | PageSuccess
    | PageError Http.Error


init : () -> ( Model, Cmd Msg )
init _ =
    ( { form = Form "" "" 150, page = MainPage }, Cmd.none )



-- UPDATE


type Msg
    = FormDataMsg FormDataMsg
    | Save
    | Done (Result Http.Error ())


type FormDataMsg
    = Name String
    | Child String
    | Amount Int


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        FormDataMsg innerMsg ->
            let
                f =
                    model.form
            in
            case innerMsg of
                Name name ->
                    ( { model | form = { f | name = name } }, Cmd.none )

                Child child ->
                    ( { model | form = { f | child = child } }, Cmd.none )

                Amount amount ->
                    ( { model | form = { f | amount = amount } }, Cmd.none )

        Save ->
            ( model
            , Http.post
                { url = "/save"
                , body = Http.jsonBody <| toJSON model.form
                , expect = Http.expectWhatever Done
                }
            )

        Done result ->
            case result of
                Ok _ ->
                    ( { model | page = PageSuccess }, Cmd.none )

                Err e ->
                    ( { model | page = PageError e }, Cmd.none )


toJSON : Form -> E.Value
toJSON f =
    E.object
        [ ( "name", E.string f.name )
        , ( "child", E.string f.child )
        , ( "amount", E.int f.amount )
        ]



-- VIEW


view : Model -> Html Msg
view model =
    div [ classes "container p-3 pb-5" ]
        [ main_ []
            [ h1 [ class "pb-5" ] [ text "Schnelle Umfrage f??r Ihren Klassenelternsprecher" ]
            , case model.page of
                MainPage ->
                    mainPage model.form

                PageSuccess ->
                    pageSuccess

                PageError e ->
                    pageError e
            ]
        ]


mainPage : Form -> Html Msg
mainPage f =
    div []
        [ p [] [ text "Liebe Eltern der Klasse 2a," ]
        , p [] [ text "danke, dass Sie sich kurz die Zeit nehmen und das folgende Formular ausf??llen." ]
        , p [ class "mb-5" ] [ text "Es gr????t Sie herzlich Norman J??ckel." ]
        , amountForm f
        ]


amountForm : Form -> Html Msg
amountForm model =
    form [ class "mb-3", onSubmit Save ]
        [ div [ class "col-md-4" ]
            [ div [ class "mb-3" ]
                [ input
                    [ class "form-control"
                    , type_ "text"
                    , placeholder "Name des Elternteils"
                    , attribute "aria-label" "Name des Elternteils"
                    , required True
                    , onInput Name
                    , value model.name
                    ]
                    []
                    |> map FormDataMsg
                ]
            , div [ class "mb-3" ]
                [ input
                    [ class "form-control"
                    , type_ "text"
                    , placeholder "Name des Kindes"
                    , attribute "aria-label" "Name des Kindes"
                    , required True
                    , onInput Child
                    , value model.child
                    ]
                    []
                    |> map FormDataMsg
                ]
            , div [ class "mb-3" ]
                [ input
                    [ class "form-control"
                    , type_ "number"
                    , placeholder "Betrag"
                    , attribute "aria-label" "Betrag"
                    , required True
                    , Html.Attributes.min "50"
                    , Html.Attributes.max "1000"
                    , onInput (String.toInt >> Maybe.withDefault 0 >> Amount)
                    , value <| String.fromInt model.amount
                    ]
                    []
                    |> map FormDataMsg
                , div [ class "form-text" ] [ text "Maximale Kosten f??r eine Klassenfahrt von 3 Tagen in Euro." ]
                ]
            , div [] [ button [ classes "btn btn-primary", type_ "submit" ] [ text "Senden" ] ]
            ]
        ]


pageSuccess : Html Msg
pageSuccess =
    div [] [ text "Das war's schon. Vielen Dank!" ]


pageError : Http.Error -> Html Msg
pageError e =
    let
        s : String
        s =
            case e of
                Http.BadUrl u ->
                    "bad url (" ++ u ++ ")"

                Http.Timeout ->
                    "timeout"

                Http.NetworkError ->
                    "network error"

                Http.BadStatus i ->
                    "bad status (" ++ String.fromInt i ++ ")"

                Http.BadBody b ->
                    "bad body (" ++ b ++ ")"
    in
    div [] [ text <| "Fehler. Versuchen Sie es noch einmal oder geben Sie auf (" ++ s ++ ")." ]


{-| This helper takes a string with class names separated by one whitespace. All
classes are applied to the result.
import Html exposing (..)
view : Model -> Html msg
view model =
div [ classes "center with-border nice-color" ][ text model.content ]
-}
classes : String -> Html.Attribute msg
classes s =
    let
        cl : List ( String, Bool )
        cl =
            String.split " " s |> List.map (\c -> ( c, True ))
    in
    Html.Attributes.classList cl
