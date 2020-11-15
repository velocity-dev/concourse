module Login.Login exposing (Model, update, userDisplayName, view)

import Concourse
import EffectTransformer exposing (ET)
import Html exposing (Html)
import Html.Attributes exposing (attribute, href, id, style, src)
import Html.Events exposing (onClick)
import Login.Styles as Styles
import Message.Effects exposing (Effect(..))
import Message.Message exposing (DomID(..), Message(..))
import UserState exposing (UserState(..))

type alias Model r =
    { r | isUserMenuExpanded : Bool }


update : Message -> ET (Model r)
update msg ( model, effects ) =
    case msg of
        Click LoginButton ->
            ( model, effects ++ [ RedirectToLogin ] )

        Click LogoutButton ->
            ( model, effects ++ [ SendLogOutRequest ] )

        Click UserMenu ->
            ( { model | isUserMenuExpanded = not model.isUserMenuExpanded }
            , effects
            )

        _ ->
            ( model, effects )


view : UserState -> Model r -> Html Message
view userState model =
    Html.div
        (id "login-component" :: Styles.loginComponent)
        (viewLoginState userState model.isUserMenuExpanded)


viewLoginState : UserState -> Bool -> List (Html Message)
viewLoginState userState isUserMenuExpanded =
    case userState of
        UserStateUnknown ->
            []

        UserStateLoggedOut ->
            [ Html.div
                ([ href "/sky/login"
                 , attribute "aria-label" "Log In"
                 , id "login-container"
                 , onClick <| Click LoginButton
                 ]
                    ++ Styles.loginContainer
                )
                [ Html.div
                    (id "login-item" :: Styles.loginItem)
                    [ Html.a
                        [ href "/sky/login" ]
                        [ Html.text "login" ]
                    ]
                ]
            ]

        UserStateLoggedIn user ->
            [ Html.div
                ([ id "login-container"
                 , onClick <| Click UserMenu
                 ]
                    ++ Styles.loginContainer
                )
                [ Html.div (id "user-id" :: Styles.loginItem)
                    (Html.div
                    [ id "user-name-and-org-container"]
                        [
                            Html.div
                                [ id "user-name-and-org"
                                , style "display" "flex"
                                , style "flex-direction" "column"
                                , style "cursor" "pointer"
                                ]
                                [ Html.span 
                                    [ style "font-size" "18px"
                                    , style "font-weight" "800"]
                                    [Html.text "Mayank Raj"]
                                , Html.span
                                    [ ]
                                    [Html.text "Velocity.Dev"]
                                ]
                        ]
                        :: (if isUserMenuExpanded then
                                [ Html.div
                                    ( [ id "org-details-expanded" ]
                                        ++ Styles.orgDetailsContainer
                                    )
                                    [ Html.div
                                        [ style "text-transform" "uppercase"
                                        , style "font-size" "20px"
                                        , style "font-weight" "800"
                                        , style "border-bottom" "2px solid #3d3c3c"]
                                        [Html.text "Organizations"]
                                    , Html.div
                                        [ id "org-list"]
                                        [ Html.div
                                            [ style "margin-top" "7px"
                                            , style "display" "flex"
                                            , style "justify-content" "start"]
                                            [ Html.img
                                                [ src "/public/images/server-logo.png"
                                                , style "height" "25px"
                                                , style "color" "#3d3c3c"]
                                                []
                                            , Html.div
                                                []
                                                [ Html.span
                                                    [ style "font-size" "18px"
                                                    , style "margin-left" "5px" 
                                                    ]
                                                    [Html.text "Velocity.Dev"]
                                                ]
                                            , Html.span
                                                [ style "margin-left" "auto"]
                                                [ Html.button
                                                    [ style "background" "transparent"
                                                    , style "color" "white"
                                                    , style "border" "1.5px solid #3d3c3c"
                                                    , style "padding" "3px 5px"
                                                    , style "font-family" "Inconsolata,monospace"
                                                    , style "border-radius" "2px"
                                                    , style "cursor" "pointer"
                                                    ]
                                                    [Html.text "Current"]
                                                , Html.button
                                                    [ style "background" "transparent"
                                                    , style "color" "white"
                                                    , style "border" "1.5px solid #3d3c3c"
                                                    , style "padding" "3px 5px"
                                                    , style "font-family" "Inconsolata,monospace"
                                                    , style "margin-left" "5px"
                                                    , style "border-radius" "2px"
                                                    , style "cursor" "pointer"
                                                    ]
                                                    [Html.text "Manage"]
                                                ]
                                            ]
                                        ]
                                    ]
                                ]
                            else
                                []
                           )
                    )
                ]
            ]


userDisplayName : Concourse.User -> String
userDisplayName user =
    Maybe.withDefault user.id <|
        List.head <|
            List.filter
                (not << String.isEmpty)
                [ "", user.userName, user.name, user.email ]
