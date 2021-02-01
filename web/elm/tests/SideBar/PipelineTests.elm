module SideBar.PipelineTests exposing (all)

import Assets
import Colors
import Common
import Concourse
import Data
import Expect
import HoverState exposing (TooltipPosition(..))
import Html exposing (Html)
import Message.Message exposing (DomID(..), Message, PipelinesSection(..))
import Set
import SideBar.Pipeline as Pipeline
import SideBar.Styles as Styles
import SideBar.Views as Views
import Test exposing (Test, describe, test)
import Test.Html.Query as Query
import Test.Html.Selector exposing (style)


defaultState =
    { active = False
    , hovered = False
    , favorited = False
    , isFavoritesSection = False
    }


all : Test
all =
    describe "sidebar pipeline"
        [ describe "when active"
            [ describe "when hovered"
                [ test "pipeline background is dark with bright border" <|
                    \_ ->
                        pipeline
                            |> viewPipeline { defaultState | active = True, hovered = True }
                            |> .background
                            |> Expect.equal Styles.Dark
                , test "pipeline icon is white" <|
                    \_ ->
                        pipeline
                            |> viewPipeline { defaultState | active = True, hovered = True }
                            |> .icon
                            |> Expect.equal Assets.PipelineIconWhite
                , describe "when not favorited"
                    [ test "displays a bright unfilled star icon when hovered" <|
                        \_ ->
                            pipeline
                                |> viewPipeline { defaultState | active = True, hovered = True }
                                |> .starIcon
                                |> Expect.equal { filled = False, isBright = True }
                    ]
                , describe "when favorited"
                    [ test "displays a bright filled star icon" <|
                        \_ ->
                            pipeline
                                |> viewPipeline
                                    { defaultState
                                        | active = True
                                        , hovered = True
                                        , favorited = True
                                    }
                                |> .starIcon
                                |> Expect.equal { filled = True, isBright = True }
                    ]
                ]
            , describe "when unhovered"
                [ test "pipeline background is dark" <|
                    \_ ->
                        pipeline
                            |> viewPipeline { defaultState | active = True, hovered = False }
                            |> .background
                            |> Expect.equal Styles.Dark
                , test "pipeline icon is bright" <|
                    \_ ->
                        pipeline
                            |> viewPipeline { defaultState | active = True, hovered = False }
                            |> .icon
                            |> Expect.equal Assets.PipelineIconLightGrey
                , describe "when unfavorited"
                    [ test "displays a dim unfilled star icon" <|
                        \_ ->
                            pipeline
                                |> viewPipeline { defaultState | active = True, hovered = False }
                                |> .starIcon
                                |> Expect.equal { filled = False, isBright = True }
                    ]
                , describe "when favorited"
                    [ test "displays a bright filled star icon" <|
                        \_ ->
                            pipeline
                                |> viewPipeline
                                    { defaultState
                                        | active = True
                                        , hovered = True
                                        , favorited = True
                                    }
                                |> .starIcon
                                |> Expect.equal { filled = True, isBright = True }
                    ]
                ]
            , test "font weight is bold" <|
                \_ ->
                    pipeline
                        |> viewPipeline { defaultState | active = True }
                        |> .name
                        |> .weight
                        |> Expect.equal Styles.Bold
            ]
        , describe "when inactive"
            [ describe "when hovered"
                [ test "pipeline background is light" <|
                    \_ ->
                        pipeline
                            |> viewPipeline { defaultState | active = False, hovered = True }
                            |> .background
                            |> Expect.equal Styles.Light
                , test "pipeline icon is white" <|
                    \_ ->
                        pipeline
                            |> viewPipeline { defaultState | active = False, hovered = True }
                            |> .icon
                            |> Expect.equal Assets.PipelineIconWhite
                ]
            , describe "when unhovered"
                [ test "pipeline name is grey" <|
                    \_ ->
                        pipeline
                            |> viewPipeline { defaultState | active = False, hovered = False }
                            |> .name
                            |> .color
                            |> Expect.equal Styles.Grey
                , test "pipeline has no background" <|
                    \_ ->
                        pipeline
                            |> viewPipeline { defaultState | active = False, hovered = False }
                            |> .background
                            |> Expect.equal Styles.Invisible
                , test "pipeline icon is dim" <|
                    \_ ->
                        pipeline
                            |> viewPipeline { defaultState | active = False, hovered = False }
                            |> .icon
                            |> Expect.equal Assets.PipelineIconGrey
                ]
            , test "font weight is default" <|
                \_ ->
                    pipeline
                        |> viewPipeline { defaultState | active = False }
                        |> .name
                        |> .weight
                        |> Expect.equal Styles.Default
            ]
        , describe "when archived"
            [ test "pipeline icon is archived" <|
                \_ ->
                    pipeline
                        |> Data.withArchived True
                        |> viewPipeline defaultState
                        |> .icon
                        |> Expect.equal Assets.ArchivedPipelineIcon
            ]
        , describe "when in all pipelines section"
            [ test "domID is for AllPipelines section" <|
                \_ ->
                    pipeline
                        |> viewPipeline { defaultState | isFavoritesSection = False }
                        |> .domID
                        |> Expect.equal
                            (SideBarPipeline AllPipelinesSection Data.pipelineId)
            ]
        , describe "when in favorites section"
            [ test "domID is for Favorites section" <|
                \_ ->
                    pipeline
                        |> viewPipeline { defaultState | isFavoritesSection = True }
                        |> .domID
                        |> Expect.equal
                            (SideBarPipeline FavoritesSection Data.pipelineId)
            ]
        ]


viewPipeline :
    { active : Bool
    , hovered : Bool
    , favorited : Bool
    , isFavoritesSection : Bool
    }
    -> Concourse.Pipeline
    -> Views.Pipeline
viewPipeline { active, hovered, favorited, isFavoritesSection } p =
    let
        hoveredDomId =
            if hovered then
                HoverState.Hovered (SideBarPipeline AllPipelinesSection Data.pipelineId)

            else
                HoverState.NoHover

        activePipeline =
            if active then
                Just Data.pipelineId

            else
                Nothing

        favoritedPipelines =
            if favorited then
                Set.singleton p.id

            else
                Set.empty
    in
    Pipeline.pipeline
        { hovered = hoveredDomId
        , currentPipeline = activePipeline
        , favoritedPipelines = favoritedPipelines
        , isFavoritesSection = isFavoritesSection
        }
        p


pipeline =
    Data.pipeline "team" 0 |> Data.withName "pipeline"


pipelineIcon : Html Message -> Query.Single Message
pipelineIcon =
    Query.fromHtml
        >> Query.children []
        >> Query.index 0


pipelineName : Html Message -> Query.Single Message
pipelineName =
    Query.fromHtml
        >> Query.children []
        >> Query.index 1
