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
    = PageOne
    | PageTwo


init : () -> ( Model, Cmd Msg )
init _ =
    ( { form = Form "" "" 150, page = PageOne }, Cmd.none )



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

        Done _ ->
            ( { model | page = PageTwo }, Cmd.none )


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
            [ h1 [ class "pb-5" ] [ text "Schnelle Umfrage für Ihren Klassenelternsprecher" ]
            , case model.page of
                PageOne ->
                    pageOne model.form

                PageTwo ->
                    pageTwo
            ]
        ]


pageOne : Form -> Html Msg
pageOne f =
    div []
        [ p [] [ text "Liebe Eltern der Klasse 2a," ]
        , p [] [ text "danke, dass Sie sich kurz die Zeit nehmen und das folgende Formular ausfüllen." ]
        , p [ class "mb-5" ] [ text "Es grüßt Sie herzlich Norman Jäckel." ]
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
                    , Html.Attributes.min "150"
                    , onInput (String.toInt >> Maybe.withDefault 0 >> Amount)
                    , value <| String.fromInt model.amount
                    ]
                    []
                    |> map FormDataMsg
                , div [ class "form-text" ] [ text "Maximale Kosten für eine Klassenfahrt von 3 Tagen in Euro." ]
                ]
            , div [] [ button [ classes "btn btn-primary", type_ "submit" ] [ text "Senden" ] ]
            ]
        ]


pageTwo : Html Msg
pageTwo =
    div [] [ text "Das war's schon. Vielen Dank!" ]


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
